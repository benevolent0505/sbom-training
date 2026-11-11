// Go の GraphQL API を呼び出すサーバーサイド専用のクライアント。

export type Post = {
  id: string;
  title: string;
  body: string;
  author: string;
  createdAt: string;
};

export type NewPost = Pick<Post, "title" | "body" | "author">;

export class ApiError extends Error {
  constructor(
    message: string,
    /** GraphQL のエラーとして API が返したメッセージ。入力エラーの表示に使う */
    readonly graphqlMessages: string[] = [],
  ) {
    super(message);
    this.name = "ApiError";
  }
}

function endpoint(): string {
  return process.env.API_URL ?? "http://localhost:8080/query";
}

export async function gql<T>(
  query: string,
  variables?: Record<string, unknown>,
): Promise<T> {
  const res = await fetch(endpoint(), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ query, variables }),
    cache: "no-store",
  });
  if (!res.ok) {
    throw new ApiError(`API responded with HTTP ${res.status}`);
  }
  const payload = (await res.json()) as {
    data?: T;
    errors?: { message: string }[];
  };
  if (payload.errors?.length) {
    const messages = payload.errors.map((e) => e.message);
    throw new ApiError(messages.join("; "), messages);
  }
  if (payload.data === undefined || payload.data === null) {
    throw new ApiError("API returned no data");
  }
  return payload.data;
}

const POST_FIELDS = "id title body author createdAt";

export async function listPosts(): Promise<Post[]> {
  const data = await gql<{ posts: Post[] }>(`{ posts { ${POST_FIELDS} } }`);
  return data.posts;
}

export async function getPost(id: string): Promise<Post | null> {
  const data = await gql<{ post: Post | null }>(
    `query($id: ID!) { post(id: $id) { ${POST_FIELDS} } }`,
    { id },
  );
  return data.post;
}

export async function createPost(input: NewPost): Promise<Post> {
  const data = await gql<{ createPost: Post }>(
    `mutation($input: NewPost!) { createPost(input: $input) { ${POST_FIELDS} } }`,
    { input },
  );
  return data.createPost;
}
