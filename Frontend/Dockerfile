# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .

# IMPORTANTE: Vite grava VITE_API_URL dentro do JS já compilado no momento do
# build — não é uma variável de ambiente normal que dá pra trocar depois só
# reiniciando o container. Se precisar apontar pra outra API, tem que
# reconstruir a imagem com --build-arg diferente.
ARG VITE_API_URL=http://localhost:8080
ENV VITE_API_URL=$VITE_API_URL

RUN npm run build

# Serve stage
FROM nginx:alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
