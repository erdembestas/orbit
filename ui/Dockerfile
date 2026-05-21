FROM node:20-alpine AS build
WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm install

COPY . .
RUN npm run build

FROM nginx:1.27-alpine
LABEL org.opencontainers.image.source="https://github.com/erdembestas/orbit"
LABEL org.opencontainers.image.description="Orbit UI image for the single-cluster Kubernetes control plane preview."
LABEL org.opencontainers.image.licenses="Apache-2.0"
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080
