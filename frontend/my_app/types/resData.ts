export type resData = {
    message: string | undefined
    // the backend sends error messages in "error" (SendError); "err" is kept for older responses
    error?: string
    err: string | undefined
    data: any | undefined
}
