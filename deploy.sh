#!/usr/bin/env bash
# ============================================
#  亲子积分系统 - Linux 部署脚本
#  用法: ./deploy.sh              # 交叉编译 + 打包
#        ./deploy.sh build        # 仅编译
#        ./deploy.sh run [port]   # 本地运行（Linux）
# ============================================
set -euo pipefail

PROJECT="credits"
OUT_DIR="deploy"
BINARY="${OUT_DIR}/${PROJECT}"
ZIPFILE="${PROJECT}-linux-amd64.tar.gz"

# ---------- 编译 ----------
build() {
    echo "==> 交叉编译 Linux amd64 ..."
    mkdir -p "${OUT_DIR}"

    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
        go build -ldflags="-s -w" -o "${BINARY}" .

    echo "==> 复制静态文件 ..."
    cp -r static "${OUT_DIR}/"

    echo ""
    echo "构建完成！输出: ${OUT_DIR}/"
    ls -lh "${BINARY}"
}

# ---------- 打包 ----------
package() {
    build
    echo ""
    echo "==> 打包为 ${ZIPFILE} ..."
    tar -czf "${ZIPFILE}" -C "${OUT_DIR}" .
    echo "打包完成: ${ZIPFILE} ($(du -h ${ZIPFILE} | cut -f1))"
    echo ""
    echo "----------------------------------------"
    echo "  部署到 Linux 服务器:"
    echo "----------------------------------------"
    echo "  1. 上传: scp ${ZIPFILE} user@host:~/"
    echo "  2. 解压: tar -xzf ${ZIPFILE} -C /opt/credits"
    echo "  3. 启动: cd /opt/credits && nohup ./${PROJECT} -port 8080 > app.log 2>&1 &"
    echo "  4. 访问: http://服务器IP:8080/child"
    echo "----------------------------------------"
}

# ---------- 本地运行（Linux 上直接跑） ----------
run() {
    local port="${1:-8080}"
    if [ ! -f "${BINARY}" ]; then
        build
    fi
    echo "==> 启动服务，端口: ${port} ..."
    cd "${OUT_DIR}" && ./"${PROJECT}" -port "${port}"
}

# ---------- systemd 服务文件 ----------
service_file() {
    cat << 'EOF'
# /etc/systemd/system/credits.service
# 部署后执行: sudo systemctl daemon-reload && sudo systemctl enable --now credits

[Unit]
Description=亲子积分任务系统
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/credits
ExecStart=/opt/credits/credits -port 8080
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF
}

# ---------- 入口 ----------
case "${1:-package}" in
    build)
        build
        ;;
    run)
        run "${2:-8080}"
        ;;
    service)
        service_file
        ;;
    package|*)
        package
        ;;
esac
