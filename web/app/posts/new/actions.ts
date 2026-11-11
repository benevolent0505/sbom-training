"use server";

import { redirect } from "next/navigation";
import { ApiError, createPost } from "../../../lib/api";

export type FormState = { error: string | null };

export async function createPostAction(
  _prev: FormState,
  formData: FormData,
): Promise<FormState> {
  const input = {
    title: String(formData.get("title") ?? ""),
    body: String(formData.get("body") ?? ""),
    author: String(formData.get("author") ?? ""),
  };

  let id: string;
  try {
    id = (await createPost(input)).id;
  } catch (e) {
    if (e instanceof ApiError && e.graphqlMessages.length > 0) {
      return { error: e.graphqlMessages.join(" / ") };
    }
    console.error("createPost failed", e);
    return { error: "記事を保存できませんでした。時間をおいて再度お試しください。" };
  }
  // redirect は例外で制御を移すため try の外で呼ぶ
  redirect(`/posts/${id}`);
}
