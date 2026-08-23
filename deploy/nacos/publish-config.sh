#!/bin/bash
# 发布配置到 Nacos
# 使用: ./publish-config.sh [service-name]

NACOS_SERVER="${NACOS_SERVER:-127.0.0.1:8848}"
NACOS_GROUP="${NACOS_GROUP:-MARKETING_GROUP}"
NACOS_NAMESPACE="${NACOS_NAMESPACE:-dev}"

SERVICE_NAME=${1:-all}

publish_config() {
    local service=$1
    local config_file="configs/${service}/config.yaml"
    local data_id="${service}.yaml"
    
    if [ -f "$config_file" ]; then
        echo "发布配置: $data_id -> $NACOS_SERVER"
        curl -X POST "http://$NACOS_SERVER/nacos/v1/cs/configs" \
            -d "dataId=$data_id&group=$NACOS_GROUP&tenant=$NACOS_NAMESPACE&content=$(cat $config_file | base64 -w0)&type=yaml" \
            -H "Content-Type: application/x-www-form-urlencoded"
        echo ""
        echo "✓ $data_id 发布成功"
    else
        echo "✗ 配置文件不存在: $config_file"
    fi
}

if [ "$SERVICE_NAME" = "all" ]; then
    for service in gateway seckill groupbuy lottery stock; do
        publish_config $service
    done
else
    publish_config $SERVICE_NAME
fi

echo ""
echo "配置发布完成！"
echo "访问 Nacos 控制台: http://$NACOS_SERVER/nacos"
