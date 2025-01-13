FROM openresty/openresty:jammy

# 安装基础依赖
RUN apt-get update && apt-get install -y \
    luarocks \
    git \
    && rm -rf /var/lib/apt/lists/*

# 安装 Lua 测试依赖
RUN luarocks install busted \
    && luarocks install luacov \
    && luarocks install lua-cjson \
    && luarocks install luasocket \
    && luarocks install lua-resty-http \
    && git clone https://github.com/knyar/nginx-lua-prometheus.git \
    && cd nginx-lua-prometheus \
    && luarocks make *.rockspec

# 设置工作目录
WORKDIR /app/waf

# 设置环境变量
ENV PATH="/usr/local/openresty/nginx/sbin:${PATH}"
ENV LUA_PATH="/app/waf/?.lua;/app/waf/test/?.lua;/usr/local/openresty/lualib/?.lua;;"

# 测试入口命令
CMD ["tail", "-f", "/dev/null"]
