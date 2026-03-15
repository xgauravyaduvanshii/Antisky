# Node.js Production Dockerfile
# Optimized for Next.js, Express, NestJS, and general Node.js apps

FROM node:20-alpine AS deps
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production && cp -R node_modules /prod_modules
RUN npm ci

FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 appuser

COPY --from=builder /app/package*.json ./
COPY --from=deps /prod_modules ./node_modules
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/.next ./.next 2>/dev/null || true
COPY --from=builder /app/public ./public 2>/dev/null || true

USER appuser

EXPOSE 3000
ENV PORT=3000

CMD ["node", "dist/index.js"]
