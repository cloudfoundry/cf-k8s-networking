FROM cloudfoundry/cflinuxfs3

COPY setup.sh .
RUN ./setup.sh
COPY roll.sh .
ENTRYPOINT ./roll.sh
