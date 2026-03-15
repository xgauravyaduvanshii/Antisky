# PHP Production Dockerfile (Laravel-optimized)
# nginx + php-fpm with OPcache

FROM composer:2 AS deps
WORKDIR /app
COPY composer.json composer.lock ./
RUN composer install --no-dev --no-scripts --no-autoloader

FROM php:8.3-fpm-alpine
RUN apk add --no-cache nginx supervisor
RUN docker-php-ext-install pdo_mysql opcache

COPY docker/nginx.conf /etc/nginx/http.d/default.conf
COPY docker/supervisord.conf /etc/supervisor/conf.d/supervisord.conf
COPY docker/php.ini /usr/local/etc/php/conf.d/custom.ini

WORKDIR /app
COPY --from=deps /app/vendor ./vendor
COPY . .

RUN composer dump-autoload --optimize 2>/dev/null || true
RUN chown -R www-data:www-data storage bootstrap/cache 2>/dev/null || true

EXPOSE 8080
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
