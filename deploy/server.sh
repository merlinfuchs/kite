#!/usr/bin/env bash
# Regenerates the nginx config and the systemd units, installs them on the
# server and reloads nginx. When the number of clusters changed, all clusters
# are restarted at once, since apps move between them. Otherwise unit changes
# take effect with the next deploy/service.sh or deploy/support.sh. The web
# server is restarted every time.
source "$(dirname "$0")/common.sh"

"$ROOT/deploy/generate.sh"

scp "$ROOT/deploy/nginx.conf" "$DEPLOY_HOST:/tmp/kite-nginx.conf"
scp "$ROOT/deploy/kite-service@.service" "$DEPLOY_HOST:/tmp/kite-service@.service"
scp "$ROOT/deploy/kite-shared.service" "$DEPLOY_HOST:/tmp/kite-shared.service"
scp "$ROOT/deploy/kite-web.service" "$DEPLOY_HOST:/tmp/kite-web.service"
scp "$ROOT/deploy/kite-support.service" "$DEPLOY_HOST:/tmp/kite-support.service"
scp "$ROOT/deploy/certbot-deploy-hook.sh" "$DEPLOY_HOST:/tmp/kite-certbot-deploy-hook.sh"
ssh -t "$DEPLOY_HOST" "set -e
  $REMOTE_SUDO

  if ls /etc/systemd/system/multi-user.target.wants/ | grep -q '^kite-service-[0-9]'; then
    echo 'The old kite-service-N units are still enabled, run deploy/move-to-opt.sh first'
    exit 1
  fi

  # The services run as their own users, which can read their config but not
  # change it or their binaries. Set up before the units, which a crash
  # restart picks up right away.
  for user in kite kite-support; do
    id -u \$user >/dev/null 2>&1 || \$S useradd --system --no-create-home --home-dir /nonexistent --shell /usr/sbin/nologin \$user
  done
  \$S chown root:kite '$SERVICE_DIR' '$SERVICE_DIR/kite.toml'
  \$S chown root:kite-support '$SUPPORT_DIR' '$SUPPORT_DIR/kite-support.toml'
  \$S chmod 750 '$SERVICE_DIR' '$SUPPORT_DIR'
  \$S chmod 640 '$SERVICE_DIR/kite.toml' '$SUPPORT_DIR/kite-support.toml'

  \$S install -m 644 /tmp/kite-service@.service /etc/systemd/system/kite-service@.service
  \$S install -m 644 /tmp/kite-shared.service /etc/systemd/system/kite-shared.service
  \$S install -m 644 /tmp/kite-web.service /etc/systemd/system/kite-web.service
  \$S install -m 644 /tmp/kite-support.service /etc/systemd/system/kite-support.service
  \$S systemctl daemon-reload

  # Running before nginx sends pages to it, and checked before anything else
  # changes, so a broken web server doesn't stop this halfway.
  \$S systemctl enable kite-web
  \$S systemctl restart kite-web
  $REMOTE_WAIT_FOR_WEB

  enabled=\$(ls /etc/systemd/system/multi-user.target.wants/ | grep '^kite-service@' | sort | xargs)
  wanted=\$(for i in \$(seq 0 $((CLUSTER_COUNT - 1))); do echo kite-service@\$i.service; done | sort | xargs)
  if [ \"\$enabled\" != \"\$wanted\" ]; then
    echo \"Clusters changed from '\$enabled' to '\$wanted', restarting all of them\"
    if [ -n \"\$enabled\" ]; then \$S systemctl disable --now \$enabled; fi
    \$S systemctl enable --now \$wanted
  fi
  # Started the first time, later unit changes take effect with the next
  # deploy/service.sh like for the clusters.
  \$S systemctl enable --now kite-shared

  # nginx doesn't start without a certificate. Until certbot has issued one,
  # see deploy/README.md, a self-signed one stands in.
  \$S install -d -m 700 '$TLS_DIR'
  if [ ! -f '$TLS_DIR/fullchain.pem' ]; then
    \$S openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:prime256v1 -nodes -days 30 \\
      -subj '/CN=$WEB_DOMAIN' -keyout '$TLS_DIR/privkey.pem' -out '$TLS_DIR/fullchain.pem' 2>/dev/null
  fi
  \$S install -d -m 755 '$ACME_DIR'
  \$S install -D -m 755 /tmp/kite-certbot-deploy-hook.sh /etc/letsencrypt/renewal-hooks/deploy/kite-nginx

  rm -f /tmp/kite-nginx.conf.prev
  if [ -f '$NGINX_CONF_PATH' ]; then \$S cp '$NGINX_CONF_PATH' /tmp/kite-nginx.conf.prev; fi
  \$S install -m 644 /tmp/kite-nginx.conf '$NGINX_CONF_PATH'
  if ! \$S nginx -t; then
    echo 'Invalid nginx config, restoring the previous one'
    if [ -f /tmp/kite-nginx.conf.prev ]; then
      \$S cp /tmp/kite-nginx.conf.prev '$NGINX_CONF_PATH'
    else
      \$S rm '$NGINX_CONF_PATH'
    fi
    exit 1
  fi

  # nginx -t doesn't catch everything, e.g. a rate limit zone with a new key.
  # nginx then keeps the old config and only logs the error, so check that new
  # workers were started and restart it otherwise.
  workers=\$(pgrep -P \"\$(systemctl show -p MainPID --value nginx)\" | sort | xargs)
  \$S systemctl reload nginx
  sleep 2
  if [ \"\$(pgrep -P \"\$(systemctl show -p MainPID --value nginx)\" | sort | xargs)\" = \"\$workers\" ]; then
    echo 'nginx rejected the reload, restarting it'
    \$S tail -n 3 /var/log/nginx/error.log
    \$S systemctl restart nginx
  fi

  rm -f /tmp/kite-nginx.conf /tmp/kite-nginx.conf.prev /tmp/kite-service@.service /tmp/kite-shared.service /tmp/kite-web.service /tmp/kite-support.service /tmp/kite-certbot-deploy-hook.sh"
