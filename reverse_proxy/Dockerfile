FROM rockylinux:8.6
RUN dnf install -y httpd
RUN dnf install -y epel-release
RUN dnf install -y redis hiredis
RUN curl -L -o /tmp/cjose-0.5.1-1.el7.centos.x86_64.rpm https://github.com/pingidentity/mod_auth_openidc/releases/download/v2.3.0/cjose-0.5.1-1.el7.centos.x86_64.rpm
RUN dnf install -y /tmp/cjose-0.5.1-1.el7.centos.x86_64.rpm
RUN curl -L -o /tmp/mod_auth_openidc-2.4.11.2-1.el8.x86_64.rpm https://github.com/zmartzone/mod_auth_openidc/releases/download/v2.4.11.2/mod_auth_openidc-2.4.11.2-1.el8.x86_64.rpm
RUN dnf install -y /tmp/mod_auth_openidc-2.4.11.2-1.el8.x86_64.rpm

CMD ["/sbin/httpd -DBACKGROUND"]