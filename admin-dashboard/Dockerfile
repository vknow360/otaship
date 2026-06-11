FROM node:24-alpine as builder

WORKDIR /app

COPY package*.json ./
RUN npm i

COPY . .

RUN npm run build

FROM node:24-alpine as runner

WORKDIR /app

COPY --from=builder /app/build ./build

USER node

EXPOSE 3000

CMD ["node", "build/index.js"]