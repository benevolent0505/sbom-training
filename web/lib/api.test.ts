import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, getPost, gql, listPosts } from "./api";

function mockFetch(body: unknown, status = 200) {
  const fn = vi.fn(
    async (_url: string | URL | Request, _init?: RequestInit) =>
      new Response(JSON.stringify(body), { status }),
  );
  vi.stubGlobal("fetch", fn);
  return fn;
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.unstubAllEnvs();
});

describe("gql", () => {
  it("API_URL に query と variables を POST する", async () => {
    vi.stubEnv("API_URL", "http://api.internal:8080/query");
    const fetchMock = mockFetch({ data: { post: null } });

    await getPost("abc");

    expect(fetchMock).toHaveBeenCalledOnce();
    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toBe("http://api.internal:8080/query");
    expect(init?.method).toBe("POST");
    expect(JSON.parse(String(init?.body)).variables).toEqual({ id: "abc" });
  });

  it("data を返す", async () => {
    const posts = [
      { id: "1", title: "t", body: "b", author: "a", createdAt: "2026-01-01T00:00:00Z" },
    ];
    mockFetch({ data: { posts } });
    await expect(listPosts()).resolves.toEqual(posts);
  });

  it("GraphQL エラーをメッセージ付きの ApiError にする", async () => {
    mockFetch({ data: null, errors: [{ message: "title is required" }] });
    const err = await gql("{ posts { id } }").catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect((err as ApiError).graphqlMessages).toEqual(["title is required"]);
  });

  it("HTTP エラーを ApiError にする", async () => {
    mockFetch({}, 502);
    await expect(gql("{ posts { id } }")).rejects.toThrow("HTTP 502");
  });
});
