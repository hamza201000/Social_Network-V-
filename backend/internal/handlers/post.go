package handler

import (
	"backend/internal/models"
	"backend/internal/services"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (userHandler *UserHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in hhhh")
	user_id := r.Context().Value("user_id").(int64)
	posts, err := userHandler.UserRepo.GetPosts(user_id)
	if err != nil {
		fmt.Println("error getposts:::", err)
		SendError(w, http.StatusInternalServerError, "could not get posts")
		return
	}
	SendJSON(w, http.StatusOK, posts)
}

func (userHandler *UserHandler) GetPostByUsername(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		SendError(w, http.StatusBadRequest, "invalid username")
		return
	}
	user_id := r.Context().Value("user_id").(int64)
	post, err := userHandler.UserRepo.GetPostByUsername(username, user_id)
	if errors.Is(err, sql.ErrNoRows) {
		SendError(w, http.StatusNotFound, "post not found")
		return
	}
	if err != nil {
		fmt.Println("error getpostbyusername:::", err)
		SendError(w, http.StatusInternalServerError, "could not get post")
		return
	}
	SendJSON(w, http.StatusOK, post)
}

func (userHandler *UserHandler) GetComments(w http.ResponseWriter, r *http.Request) {
	postID, err := pathID(r)
	if err != nil {
		SendError(w, http.StatusBadRequest, "invalid post id")
		return
	}
	user_id := r.Context().Value("user_id").(int64)
	allowed, err := userHandler.UserRepo.CanViewPost(postID, user_id)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "could not check post access")
		return
	}
	if !allowed {
		SendError(w, http.StatusNotFound, "post not found")
		return
	}
	comments, err := userHandler.UserRepo.GetComments(postID)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "could not get comments")
		return
	}
	SendJSON(w, http.StatusOK, models.Res{Message: "successfully retrieved comments", Data: comments})
}

func (userHandler *UserHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	// userID, err := userHandler.requiredUserID(r)
	user_id := r.Context().Value("user_id").(int64)
	post, err := parsePostRequest(w, r)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := services.ValidatePostRequest(post); err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	if post.Privacy == "private" {
		if len(post.FollowerIDs) == 0 {
			SendError(w, http.StatusBadRequest, "private posts require selected followers")
			return
		}
		if err := userHandler.UserRepo.ValidatePrivateFollowers(user_id, post.FollowerIDs); err != nil {
			SendError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	postID, err := userHandler.UserRepo.CreatePost(user_id, post)
	if err != nil {
		fmt.Println("error createpost:::", err)
		SendError(w, http.StatusInternalServerError, "could not create post")
		return
	}
	SendJSON(w, http.StatusCreated, models.Res{Message: "post created successfully", Data: map[string]int64{"id": postID}})
}

func (userHandler *UserHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	user_id := r.Context().Value("user_id").(int64)
	postID, err := pathID(r)
	if err != nil {
		SendError(w, http.StatusBadRequest, "invalid post id")
		return
	}
	allowed, err := userHandler.UserRepo.CanViewPost(postID, user_id)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "could not check post access")
		return
	}
	if !allowed {
		SendError(w, http.StatusNotFound, "post not found")
		return
	}
	comment, err := parseCommentRequest(w, r)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := services.ValidateCommentRequest(comment); err != nil {
		SendError(w, http.StatusBadRequest, err.Error())
		return
	}
	commentID, err := userHandler.UserRepo.CreateComment(user_id, postID, comment)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "could not create comment")
		return
	}
	SendJSON(w, http.StatusCreated, models.Res{Message: "comment created successfully", Data: map[string]int64{"id": commentID}})
}

// func (userHandler *UserHandler) requiredUserID(r *http.Request) (int64, error) {
// 	cookie, err := r.Cookie("session_id")
// 	if err != nil {
// 		return 0, err
// 	}
// 	session, err := userHandler.UserRepo.GetSession(cookie.Value)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return int64(session.User_id), nil
// }

// func (userHandler *UserHandler) optionalUserID(r *http.Request) int64 {
// 	userID, err := userHandler.requiredUserID(r)
// 	if err != nil {
// 		return 0
// 	}
// 	return userID
// }

func pathID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func parsePostRequest(w http.ResponseWriter, r *http.Request) (models.PostRequest, error) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		var post models.PostRequest
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			return post, errors.New("invalid request body")
		}
		return post, nil
	}
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)
	if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil {
		return models.PostRequest{}, errors.New("request is too large or invalid")
	}
	post := models.PostRequest{Content: strings.TrimSpace(r.FormValue("content")), Privacy: strings.TrimSpace(r.FormValue("privacy"))}
	if post.Privacy == "" {
		post.Privacy = "public"
	}
	for _, value := range strings.Split(r.FormValue("follower_ids"), ",") {
		if strings.TrimSpace(value) == "" {
			continue
		}
		followerID, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil || followerID <= 0 {
			return post, errors.New("follower_ids must contain valid user ids")
		}
		post.FollowerIDs = append(post.FollowerIDs, followerID)
	}
	var err error
	post.Image, err = saveUploadedImage(r, "image")
	return post, err
}

func parseCommentRequest(w http.ResponseWriter, r *http.Request) (models.CommentRequest, error) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		var comment models.CommentRequest
		if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
			return comment, errors.New("invalid request body")
		}
		return comment, nil
	}
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)
	if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil {
		return models.CommentRequest{}, errors.New("request is too large or invalid")
	}
	comment := models.CommentRequest{Content: strings.TrimSpace(r.FormValue("content"))}
	var err error
	comment.Image, err = saveUploadedImage(r, "image")
	return comment, err
}

func saveUploadedImage(r *http.Request, field string) (string, error) {
	file, header, err := r.FormFile(field)
	if errors.Is(err, http.ErrMissingFile) {
		return "", nil
	}
	if err != nil {
		return "", errors.New("invalid image")
	}
	defer file.Close()
	if header.Size > 5*1024*1024 {
		return "", errors.New("image must be less than 5 MB")
	}
	contentType := header.Header.Get("Content-Type")
	allowedTypes := map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true, "image/gif": true}
	if !allowedTypes[contentType] {
		return "", errors.New("only JPEG, PNG, WebP and GIF images are allowed")
	}
	if err := os.MkdirAll("uploads", 0755); err != nil {
		return "", err
	}
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), filepath.Ext(header.Filename))
	path := filepath.Join("uploads", filename)
	output, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer output.Close()
	if _, err := io.Copy(output, file); err != nil {
		return "", err
	}
	return path, nil
}


func (userHandler *UserHandler) GetPostMyPosts(w http.ResponseWriter, r *http.Request) {
	
	user_id := r.Context().Value("user_id").(int64)
	post, err := userHandler.UserRepo.GetPostMyPosts(user_id)
	if errors.Is(err, sql.ErrNoRows) {
		SendError(w, http.StatusNotFound, "post not found")
		return
	}
	if err != nil {
		fmt.Println("error getpostmyposts:::", err)
		SendError(w, http.StatusInternalServerError, "could not get post")
		return
	}
	SendJSON(w, http.StatusOK, post)
}