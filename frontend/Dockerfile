FROM --platform=$BUILDPLATFORM node:22-alpine AS frontend-builder

WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM node:22-alpine AS frontend-runtime

WORKDIR /app
ENV NODE_ENV=production
COPY --from=frontend-builder /app/.next ./.next
COPY --from=frontend-builder /app/public ./public
COPY --from=frontend-builder /app/package.json ./package.json
COPY --from=frontend-builder /app/package-lock.json ./package-lock.json
RUN npm ci --omit=dev
EXPOSE 3000
CMD ["npm", "run", "start"]
