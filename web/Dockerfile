# syntax=docker/dockerfile:1

FROM node:26.12.1-trixie-slim@sha256:94816674c62774da0d099a947bf4613b9e1972a2733abd0d0cd80e795f8411d0 AS deps
WORKDIR /app
COPY package.json package-lock.json .npmrc ./
RUN --mount=type=cache,target=/root/.npm npm ci

FROM deps AS build
COPY . .
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

# 実行イメージには standalone 出力(本番で使う依存のみ)だけを入れる。
# devDependencies やビルドツールはイメージにも SBOM にも含まれない。
FROM gcr.io/distroless/nodejs24-debian13:nonroot@sha256:088a01637f22908632ba09018be2e0b516f2fa6f7869c191d29491139db43c78
WORKDIR /app
ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    PORT=3000 \
    HOSTNAME=0.0.0.0
COPY --from=build /app/.next/standalone ./
COPY --from=build /app/.next/static ./.next/static
EXPOSE 3000
CMD ["server.js"]
