"use client";

import { useActionState } from "react";
import { createPostAction, type FormState } from "./actions";

const initialState: FormState = { error: null };

export function PostForm() {
  const [state, formAction, pending] = useActionState(
    createPostAction,
    initialState,
  );

  return (
    <form action={formAction} className="form">
      <label>
        タイトル
        <input name="title" required maxLength={200} />
      </label>
      <label>
        著者
        <input name="author" required maxLength={100} />
      </label>
      <label>
        本文
        <textarea name="body" required rows={10} />
      </label>
      {state.error && (
        <p className="error" role="alert">
          {state.error}
        </p>
      )}
      <div>
        <button type="submit" className="button" disabled={pending}>
          {pending ? "保存中…" : "公開する"}
        </button>
      </div>
    </form>
  );
}
