// ALB のヘルスチェック用。API の状態には依存させない。
export function GET() {
  return Response.json({ status: "ok" });
}
