#!/usr/bin/env ruby

require 'json'
require 'yaml'
require 'base64'

puts "# Generating values to use with helm template"

if ARGV.length == 0
    puts "Usage: ./generate_values.rb [PATH_TO_BBL_STATE_FILE]"
    exit
end

# Path to bbl state
ARGV[0]

bbl_state = JSON.parse(File.read(ARGV[0]))

director_name = bbl_state['bosh']['directorName']
lb_cert_ca = bbl_state['lb']['cert']
domain = bbl_state['lb']['domain']

client_secret = JSON.parse(`credhub get -n /#{director_name}/cf/uaa_clients_network_policy_secret -j`)['value']
client_name = "network-policy"
cc_base_url = "https://api.#{domain}"
uaa_base_url = "https://uaa.#{domain}"

puts YAML.dump({
        'cfroutesync' => {
          'ccCA' => Base64.strict_encode64(lb_cert_ca),
          'ccBaseURL' => Base64.strict_encode64(cc_base_url),
          'uaaCA' => Base64.strict_encode64(lb_cert_ca),
          'uaaBaseURL' => Base64.strict_encode64(uaa_base_url),
          'clientName' => Base64.strict_encode64(client_name),
          'clientSecret' => Base64.strict_encode64(client_secret),
        }
  })
