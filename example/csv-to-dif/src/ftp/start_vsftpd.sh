mkdir -p /ftp/pub/data
chown -R ftp:ftp /ftp
chmod a-w /ftp/pub
chcon -R -t public_content_rw_t /ftp/pub
echo "pasv_address=${TURBO_ADDRESS}" >> /etc/vsftpd/vsftpd.conf

&>/dev/null /usr/sbin/vsftpd /etc/vsftpd/vsftpd.conf
