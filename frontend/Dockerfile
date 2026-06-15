FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./
RUN npm ci

# Copy source code
COPY . .

# Build Next.js
ARG API_URL=http://backend:8080/api
ENV API_URL=$API_URL
RUN npm run build

# Production runner
FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production

# Copy built assets
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

ARG PORT=3000
ENV PORT=$PORT
EXPOSE $PORT
ENV HOSTNAME="0.0.0.0"

# Note: Next.js standalone output runs server.js
CMD ["node", "server.js"]
