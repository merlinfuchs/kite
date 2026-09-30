#!/usr/bin/env bash
# Regenerates the nginx config and the systemd unit, installs them on the
# server and reloads nginx. When the number of clusters changed, all clusters
# are restarted at once, since apps move between them. Otherwise unit changes
# take effect with the next deploy/service.sh.
source "$(dirname "$0")/common.sh"

"$ROOT/deploy/generate.sh"

scp "$ROOT/deploy/nginx.conf" "$DEPLOY_HOST:/tmp/kite-nginx.conf"
scp "$ROOT/deploy/kite-service@.service" "$DEPLOY_HOST:/tmp/kite-service@.service"
ssh -t "$DEPLOY_HOST" "set -e
  $REMOTE_SUDO

  if ls /etc/systemd/system/multi-user.target.wants/ | grep -q '^kite-service-[0-9]'; then
    echo 'The old kite-service-N units are still enabled, run deploy/move-to-opt.sh first'
    exit 1
  fi

  \$S install -m 644 /tmp/kite-service@.service /etc/systemd/system/kite-service@.service
  \$S systemctl daemon-reload

  enabled=\$(ls /etc/systemd/system/multi-user.target.wants/ | grep '^kite-service@' | sort | xargs)
  wanted=\$(for i in \$(seq 0 $((CLUSTER_COUNT - 1))); do echo kite-service@\$i.service; done | sort | xargs)
  if [ \"\$enabled\" != \"\$wanted\" ]; then
    echo \"Clusters changed from '\$enabled' to '\$wanted', restarting all of them\"
    if [ -n \"\$enabled\" ]; then \$S systemctl disable --now \$enabled; fi
    \$S systemctl enable --now \$wanted
  fi

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
  \$S systemctl reload nginx

  rm -f /tmp/kite-nginx.conf /tmp/kite-nginx.conf.prev /tmp/kite-service@.service"
