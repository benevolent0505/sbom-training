import { notFound } from "next/navigation";
import { getPost } from "../../../lib/api";
import { formatDate } from "../../../lib/format";

export const dynamic = "force-dynamic";

export default async function PostPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const post = await getPost(id);
  if (!post) {
    notFound();
  }

  return (
    <article>
      <h1>{post.title}</h1>
      <p className="meta">
        {post.author}・{formatDate(post.createdAt)}
      </p>
      <p className="post-body">{post.body}</p>
    </article>
  );
}
