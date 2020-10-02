FROM cloudfoundry/cflinuxfs3

COPY setup.sh .
RUN ./setup.sh
COPY save.sh .
ENTRYPOINT ["./save.sh"]
