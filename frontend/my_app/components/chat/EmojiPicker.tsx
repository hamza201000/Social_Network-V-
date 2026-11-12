"use client";

import { useEffect, useRef, useState } from "react";
import "./emoji-picker.css";

const EMOJIS = [
  "😀", "😂", "🤣", "😊", "😍", "😘", "😎", "🤩",
  "🙂", "😉", "😅", "😇", "🤔", "😴", "😢", "😭",
  "😡", "😱", "🥳", "🤗", "🙄", "😬", "🤯", "🥺",
  "👍", "👎", "👏", "🙌", "🙏", "💪", "👋", "🤝",
  "❤️", "💔", "🔥", "✨", "🎉", "💯", "✅", "❌",
  "☕", "🍕", "⚽", "🎮", "🎵", "📚", "🌞", "🌙",
];

type EmojiPickerProps = {
  onPick: (emoji: string) => void;
};

// A small button that opens a grid of emojis; picking one calls onPick.
export default function EmojiPicker({ onPick }: EmojiPickerProps) {
  const [open, setOpen] = useState(false);
  const wrapper = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    function close(event: MouseEvent | KeyboardEvent) {
      if (event instanceof KeyboardEvent) {
        if (event.key === "Escape") setOpen(false);
        return;
      }
      if (!wrapper.current?.contains(event.target as Node)) setOpen(false);
    }
    document.addEventListener("mousedown", close);
    document.addEventListener("keydown", close);
    return () => {
      document.removeEventListener("mousedown", close);
      document.removeEventListener("keydown", close);
    };
  }, [open]);

  return (
    <div className="emoji-picker" ref={wrapper}>
      <button
        type="button"
        className="emoji-picker-toggle"
        aria-label="Add an emoji"
        aria-expanded={open}
        onClick={() => setOpen((value) => !value)}
      >
        😊
      </button>
      {open && (
        <div className="emoji-picker-panel" role="dialog" aria-label="Emojis">
          {EMOJIS.map((emoji) => (
            <button
              key={emoji}
              type="button"
              className="emoji-picker-item"
              onClick={() => onPick(emoji)}
            >
              {emoji}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
