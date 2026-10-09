#!/usr/bin/env bash
# One-off move of the hand-made setup in /root to the layout these scripts
# expect. Copies kite-service, kite-support and nirn-proxy with their configs
# to /opt, points the backup cron jobs and the kite-support and nirn-proxy
# units there, then replaces the kite-service-N units with kite-service@N
# through deploy/server.sh. The files in /root are left as they are.
source "$(dirname "$0")/common.sh"

ssh -t "$DEPLOY_HOST" "set -e
  $REMOTE_SUDO

  for dir in '$SERVICE_DIR' '$SUPPORT_DIR' '$NIRN_PROXY_DIR'; do
    if [ -e \"\$dir\" ]; then
      echo \"\$dir already exists, the move has already been done\"
      exit 1
    fi
  done

  \$S install -d -m 700 '$SERVICE_DIR' '$SUPPORT_DIR' '$NIRN_PROXY_DIR'
  \$S install -d '$WEB_DIR' '$DOCS_DIR'

  \$S cp -p /root/kite-service /root/kite.toml '$SERVICE_DIR/'
  for script in backup_daily.sh backup_weekly.sh; do
    { echo 'cd $SERVICE_DIR'; cat /root/\$script; } | \$S tee '$SERVICE_DIR'/\$script >/dev/null
  done
  \$S crontab -l | sed 's|/root/backup_|$SERVICE_DIR/backup_|' | \$S crontab -

  \$S cp -p /root/kite-support /root/kite-support.toml '$SUPPORT_DIR/'
  \$S cp -p /root/nirn-proxy '$NIRN_PROXY_DIR/'
  \$S sed -i 's|^WorkingDirectory=/root\$|WorkingDirectory=$SUPPORT_DIR|; s|^ExecStart=/root/|ExecStart=$SUPPORT_DIR/|' /etc/systemd/system/kite-support.service
  \$S sed -i 's|^WorkingDirectory=/root\$|WorkingDirectory=$NIRN_PROXY_DIR|; s|^ExecStart=/root/|ExecStart=$NIRN_PROXY_DIR/|' /etc/systemd/system/nirn-proxy.service
  \$S systemctl daemon-reload
  \$S systemctl restart kite-support nirn-proxy

  # Same config under the new name, so deploy/server.sh can restore it if the
  # new one is invalid.
  if [ -L /etc/nginx/sites-enabled/api.kite.onl ]; then
    \$S cp /etc/nginx/sites-available/api.kite.onl '$NGINX_CONF_PATH'
    \$S rm /etc/nginx/sites-enabled/api.kite.onl
  fi

  old_units=\$(ls /etc/systemd/system/multi-user.target.wants/ | grep '^kite-service-[0-9]' | xargs)
  if [ -n \"\$old_units\" ]; then
    echo \"Stopping \$old_units\"
    \$S systemctl disable --now \$old_units
  fi"

"$ROOT/deploy/server.sh"
