FROM node:20-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci

CMD ["npm", "start", "--", "--host", "0.0.0.0", "--port", "4200", "--poll", "2000"]
