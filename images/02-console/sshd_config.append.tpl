
{{- if .Port}}
Port {{.Port}}
{{- end}}

{{- if .ListenAddress}}
ListenAddress {{.ListenAddress}}
{{- end}}

ClientAliveInterval 180

UseDNS no

AllowGroups docker

# Enforce security settings
Protocol 2
PermitRootLogin no
MaxAuthTries 4
IgnoreRhosts yes
HostbasedAuthentication no
PermitEmptyPasswords no
AllowTcpForwarding no
