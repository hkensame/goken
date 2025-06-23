# 开发专用镜像
FROM rust:1.75

WORKDIR /app
COPY . .

# 安装开发工具
RUN cargo install cargo-watch && \
    rustup component add rustfmt clippy

# 启动热重载监控
CMD ["cargo", "watch", "-x", "run"]