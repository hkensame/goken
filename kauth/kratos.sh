docker run --rm \
  -v "$(pwd)/configs/kratos/kratos.yml:/etc/kratos/config.yaml" \
  oryd/kratos:latest \
  migrate sql -e --yes --config /etc/kratos/config.yaml

docker run -d --name kratos \
  -p 4433:4433 -p 4434:4434 \
  -v "$(pwd)/configs/kratos/kratos.yml:/etc/kratos/config.yaml" \
  -v "$(pwd)/configs/kratos/identity.schema.json:/etc/kratos/identity.schema.json" \
  oryd/kratos:latest \
  serve --dev --config /etc/kratos/config.yaml
