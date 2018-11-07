ctrl_interface=/var/run/wpa_supplicant
ap_scan=1
update_config=1

{{- range $key, $value := .}}
network={
	ssid="{{$value.SSID}}"
	{{- if gt (len $value.PSK) 0}}
	psk="{{$value.PSK}}"
	{{- end}}
	{{- if gt (len $value.KeyMgmt) 0}}
	key_mgmt={{$value.KeyMgmt}}
	{{- end}}
	{{- if $value.ScanSSID}}
	scan_ssid={{$value.ScanSSID}}
	{{- end}}
	{{- if $value.Priority}}
	priority={{$value.Priority}}
	{{- end}}
	{{- if gt (len $value.Pairwise) 0}}
	pairwise={{$value.Pairwise}}
	{{- end}}
	{{- if gt (len $value.Group) 0}}
	group={{$value.Group}}
	{{- end}}
	{{- if gt (len $value.Eap) 0}}
	eap={{$value.Eap}}
	{{- end}}
	{{- if gt (len $value.Identity) 0}}
	identity="{{$value.Identity}}"
	{{- end}}
	{{- if gt (len $value.AnonymousIdentity) 0}}
	anonymous_identity="{{$value.AnonymousIdentity}}"
	{{- end}}
	{{- if $value.EapolFlags}}
	eapol_flags={{$value.EapolFlags}}
	{{- end}}
	{{- if gt (len $value.Password) 0}}
	password="{{$value.Password}}"
	{{- end}}
	{{- range $i, $v := $value.Phases}}
	phase{{addFunc $i 1}}="{{$v}}"
	{{- end}}
	{{- range $i, $v := $value.CaCerts}}
	{{- if eq $i 0}}
	ca_cert="{{$v}}"
	{{- else}}
	ca_cert{{addFunc $i 1}}="{{$v}}"
	{{- end}}
	{{- end}}
	{{- range $i, $v := $value.ClientCerts}}
	{{- if eq $i 0}}
	client_cert="{{$v}}"
	{{- else}}
	client_cert{{addFunc $i 1}}="{{$v}}"
	{{- end}}
	{{- end}}
	{{- range $i, $v := $value.PrivateKeys}}
	{{- if eq $i 0}}
	private_key="{{$v}}"
	{{- else}}
	private_key{{addFunc $i 1}}="{{$v}}"
	{{- end}}
	{{- end}}
	{{- range $i, $v := $value.PrivateKeyPasswds}}
	{{- if eq $i 0}}
	private_key_passwd="{{$v}}"
	{{- else}}
	private_key{{addFunc $i 1}}_passwd="{{$v}}"
	{{- end}}
	{{- end}}
}
{{- end}}