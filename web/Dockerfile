# Stage 1: Install development dependencies
FROM node:20-alpine AS development-dependencies-env
WORKDIR /app
COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile

# Stage 2: Install only production dependencies
FROM node:20-alpine AS production-dependencies-env
WORKDIR /app
COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile --production

# Stage 3: Build app
FROM node:20-alpine AS build-env
WORKDIR /app

ARG VITE_FLX_INTERNAL_URL
ARG VITE_FLX_API_URL

ENV VITE_FLX_INTERNAL_URL=$VITE_FLX_INTERNAL_URL
ENV VITE_FLX_API_URL=$VITE_FLX_API_URL

COPY . .
COPY --from=development-dependencies-env /app/node_modules ./node_modules
RUN yarn build

# Stage 4: Final image
FROM node:20-alpine
WORKDIR /app
COPY package.json yarn.lock ./
COPY --from=production-dependencies-env /app/node_modules ./node_modules
COPY --from=build-env /app/build ./build
EXPOSE 3000
CMD ["yarn", "start"]
