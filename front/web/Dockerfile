# Build context: front/web. The static build is the same for every platform: Node runs on the
# builder's own one and only the nginx stage is multi-arch.
FROM --platform=$BUILDPLATFORM node:22-alpine AS build
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
# Public values baked into the bundle (empty backend url: same origin as the page).
ARG VITE_BACKEND_URL=
ARG VITE_DROPBOX_API_KEY=
ARG VITE_GOOGLE_WEB_CLIENT_ID=
RUN npm run build

FROM nginxinc/nginx-unprivileged:1.27-alpine AS runtime
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /src/dist /usr/share/nginx/html
EXPOSE 8080
