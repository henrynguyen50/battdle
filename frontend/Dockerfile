FROM nginx:alpine

# Remove default configuration and copy custom one
RUN rm /etc/nginx/conf.d/default.conf
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy static frontend files to Nginx html directory
COPY . /usr/share/nginx/html/

EXPOSE 3000
