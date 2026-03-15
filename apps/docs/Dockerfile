FROM node:20-alpine AS builder
# Set working directory to the app folder
WORKDIR /app
# Copy the package.json of the docs app
COPY apps/docs/package.json ./
# Install dependencies
RUN npm install
# Copy the rest of the docs app source code
COPY apps/docs/ ./
# Copy the markdown files to the root directory so the path ../../README.md resolves correctly
COPY docs/ /docs/
COPY README.md /README.md
# Build the Next.js app
RUN npm run build

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./
COPY --from=builder /docs /docs
COPY --from=builder /README.md /README.md
EXPOSE 3002
CMD ["npm", "run", "start", "--", "-p", "3002"]
