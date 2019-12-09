#!/usr/bin/env ruby

require 'json'
require 'base64'
require 'yaml'

puts "#! Generating values to use with YTT/helm template"

if ARGV.length == 0
    puts "Usage: ./generate_values.rb [PATH_TO_BBL_STATE_FILE]"
    puts "Make sure you load CREDHUB env vars via bbl print-env prior to running"
    exit
end

bbl_state_file = ARGV[0]
bbl_state_dir = File.dirname(bbl_state_file)

unless ENV['CREDHUB_PROXY']
 puts "CREDHUB_PROXY must be set. Have you run 'eval \"$(bbl print-env --state-dir=#{bbl_state_dir})\"'?"
 exit 1
end

# Path to bbl state
bbl_state = JSON.parse(File.read(bbl_state_file))

director_name = bbl_state['bosh']['directorName']
lb_cert_ca = bbl_state['lb']['cert']
domain = bbl_state['lb']['domain']

client_secret = JSON.parse(`credhub get -n /#{director_name}/cf/uaa_clients_network_policy_secret -j`)['value']
client_name = "network-policy"
cc_base_url = "https://api.#{domain}"
uaa_base_url = "https://uaa.#{domain}"

# for Eirini 1.0+ set this to cloudfoundry.org/
eirini_pod_label_prefix = "cloudfoundry.org/"

puts '#@data/values'
puts YAML.dump({
        'cfroutesync' => {
          'ccCA' => lb_cert_ca,
          'ccBaseURL' => cc_base_url,
          'uaaCA' => lb_cert_ca,
          'uaaBaseURL' => uaa_base_url,
          'clientName' => client_name,
          'clientSecret' => Base64.strict_encode64(client_secret),
          'eiriniPodLabelPrefix' => eirini_pod_label_prefix,
        }
  })
