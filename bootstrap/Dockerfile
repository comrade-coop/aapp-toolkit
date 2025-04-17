FROM node:18-alpine

RUN apk add --no-cache tar

WORKDIR /app

COPY package*.json ./
RUN if [ -f package.json ]; then npm install; else npm install tar; fi

COPY server.js .

EXPOSE 54321

CMD ["node", "server.js"]