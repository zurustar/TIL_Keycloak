FROM httpd:2.4
RUN apt-get update -y
RUN mkdir /usr/local/apache2/conf/include/ && echo 'Include conf/include/*.conf ' >> /usr/local/apache2/conf/httpd.conf
RUN apt-get install -y libapache2-mod-auth-openidc
COPY ./conf/*.conf   /usr/local/apache2/conf/include/
