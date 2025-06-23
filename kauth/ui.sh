docker run -d --name kratos-ui \
  -p 3000:3000 \
  -e KRATOS_PUBLIC_URL=http://192.168.199.131:4433/ \
  -e KRATOS_BROWSER_URL=http://192.168.199.131:3000/ \
  -e KRATOS_ADMIN_URL=http://192.168.199.131:4434/ \
  -e COOKIE_SECRET=verysecretcookie \
  -e CSRF_COOKIE_NAME=__Host-kratos_csrf \
  -e CSRF_COOKIE_SECRET=verysecretcsrf \
  -e PORT=3000 \
  oryd/kratos-selfservice-ui-node

# -e HYDRA_ADMIN_URL=http://192.168.199.128:4445/ \
  # -e COOKIE_SECRET=verysecretcookie \
  # -e CSRF_COOKIE_NAME=__Host-kratos_csrf \
  # -e CSRF_COOKIE_SECRET=verysecretcsrf \