FROM openresty/openresty:alpine

# 安装opm所需的依赖
RUN apk add --no-cache curl perl

# 安装额外的 Lua 模块
RUN /usr/local/openresty/bin/opm get \
    ledgetech/lua-resty-http \
    hamishforbes/lua-resty-iputils \
    knyar/nginx-lua-prometheus

# 创建WAF所需目录
RUN mkdir -p /usr/local/openresty/nginx/logs/waf && \
    mkdir -p /usr/local/openresty/waf && \
    mkdir -p /usr/local/openresty/nginx/conf/conf.d

# 设置工作目录
WORKDIR /usr/local/openresty/nginx

# 将 WAF 核心代码复制到镜像中
COPY core /usr/local/openresty/xwaf

# 复制 Nginx 配置文件
COPY core/conf/nginx.conf /usr/local/openresty/nginx/conf/nginx.conf
COPY core/conf/conf.d/*.conf /usr/local/openresty/nginx/conf/conf.d/

# 复制HTML模板文件
COPY core/html/block.html /usr/local/openresty/nginx/html/

# 启动 Nginx
CMD ["/usr/local/openresty/bin/openresty", "-g", "daemon off;"]