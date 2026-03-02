# Stage 1: Build
FROM guergeiro/pnpm:20-10 AS builder
WORKDIR /app

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm build


# Stage 2: Production runtime
FROM node:20-alpine
WORKDIR /app

# copy only the built output
COPY --from=builder /app/.output ./.output
COPY --from=builder /app/package.json ./package.json
EXPOSE 3000

CMD ["node", ".output/server/index.mjs"]
