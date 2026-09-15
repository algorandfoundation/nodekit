# Throwaway Debian container for the statewalker journey scenario.
#
# systemd runs as PID 1 so the algorand apt package and nodekit manage the
# node exactly like on an end user's machine. statewalker starts it with:
#
#   docker run -d --privileged --cgroupns=host \
#     -v /sys/fs/cgroup:/sys/fs/cgroup:rw \
#     --add-host=host.docker.internal:host-gateway \
#     -p 127.0.0.1:0:8080 statewalker-journey
FROM debian:12

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    systemd systemd-sysv dbus \
    ca-certificates curl gnupg2 sudo procps \
    && rm -rf /var/lib/apt/lists/*

# Drop units that are useless in a container and slow down boot
RUN systemctl mask \
    systemd-udevd.service systemd-udevd-kernel.socket systemd-udevd-control.socket \
    systemd-networkd.service systemd-logind.service getty.target

# The journey runs as a regular user with passwordless sudo, mirroring a real
# end-user machine (nodekit shells out to `sudo apt-get ...`). Without
# NOPASSWD, sudo would prompt for a password and the whole flow would stall
# silently at 0% CPU. env_keep forwards the noninteractive apt/debconf
# frontends through sudo so no package operation can prompt either.
RUN useradd -m -s /bin/bash algo \
    && printf 'algo ALL=(ALL) NOPASSWD:ALL\nDefaults:algo env_keep += "DEBIAN_FRONTEND APT_LISTCHANGES_FRONTEND"\n' > /etc/sudoers.d/algo \
    && chmod 0440 /etc/sudoers.d/algo \
    && visudo -cf /etc/sudoers.d/algo

# Tape recording layer (used by `statewalker stage tape`): a pinned vhs deb
# plus its terminal/render dependencies. The versions are pinned deliberately;
# the latest host-side vhs silently rendered zero frames, so upgrades must be
# revalidated against the gif sanity checks in internal/tape.go. ttyd ships as
# a static binary because Debian 12 does not package it (Ubuntu does).
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg libnss3 fonts-jetbrains-mono \
    && rm -rf /var/lib/apt/lists/* \
    && curl -fsSL -o /usr/local/bin/ttyd https://github.com/tsl0922/ttyd/releases/download/1.7.7/ttyd.x86_64 \
    && chmod +x /usr/local/bin/ttyd \
    && ttyd --version \
    && curl -fsSL -o /tmp/vhs.deb https://github.com/charmbracelet/vhs/releases/download/v0.9.0/vhs_0.9.0_amd64.deb \
    && dpkg -i /tmp/vhs.deb \
    && rm /tmp/vhs.deb \
    && vhs --version

# Never let dpkg stop on a conffile question during upgrades
RUN printf 'Dpkg::Options {\n  "--force-confdef";\n  "--force-confold";\n};\n' \
    > /etc/apt/apt.conf.d/90statewalker-noninteractive \
    && echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections

STOPSIGNAL SIGRTMIN+3

CMD ["/lib/systemd/systemd"]
