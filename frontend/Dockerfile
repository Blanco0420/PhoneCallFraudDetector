FROM node:22 AS base

WORKDIR /app

COPY ./package.json yarn.lock ./
RUN yarn

# === === DEV === ===
FROM base AS dev
CMD ["yarn", "dev", "--host"]

COPY . .

# === === PROD === ===
FROM base AS prodbuild
RUN yarn build

FROM nginx:alpine AS prod
COPY --from=prodbuild /app/dist /usr/share/nginx/html
