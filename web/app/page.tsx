import Link from "next/link";
import { listPosts } from "../lib/api";
import { formatDate } from "../lib/format";

// 記事はリクエストごとに API から取得する(ビルド時に API を必要としない)
export const dynamic = "force-dynamic";

export default async function HomePage() {
  const posts = await listPosts();

  if (posts.length === 0) {
    return (
      <p>
        まだ記事がありません。<Link href="/posts/new">最初の記事を書く</Link>
      </p>
    );
  }

  return (
    <ul className="post-list">
      {posts.map((post) => (
        <li key={post.id} className="post-card">
          <h2>
            <Link href={`/posts/${post.id}`}>{post.title}</Link>
          </h2>
          <p className="meta">
            {post.author}・{formatDate(post.createdAt)}
          </p>
        </li>
      ))}
    </ul>
  );
}
