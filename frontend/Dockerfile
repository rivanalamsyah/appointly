# ==============================================================================
# Appointly Frontend Production Dockerfile (Multi-Stage)
# ==============================================================================

# --- Stage 1: Build Stage ---
FROM node:22-alpine AS builder

WORKDIR /app

# Install dependencies with frozen lockfile
COPY package.json package-lock.json ./
RUN npm ci

# Copy source files
COPY . .

# Build Astro standalone Node server bundle
ENV NODE_ENV=production
RUN npm run build

# --- Stage 2: Production Node Runner ---
FROM node:22-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=4321

# Create non-root system user
RUN addgroup -g 10001 -S appointly && \
    adduser -u 10001 -S appointly -G appointly

# Copy standalone production build and runtime packages
COPY --from=builder /app/dist /app/dist
COPY --from=builder /app/node_modules /app/node_modules
COPY --from=builder /app/package.json /app/package.json

RUN chown -R appointly:appointly /app

USER appointly:appointly

EXPOSE 4321

# Healthcheck
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:4321/ || exit 1

CMD ["node", "./dist/server/entry.mjs"]
