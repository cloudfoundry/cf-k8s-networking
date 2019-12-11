# Metrics

## View

1. Port forward grafana pod's port:

   ```bash
   istioctl dashboard grafana
   ```
   
1. Go to "Indicator" dashboard.

## Export

1. Go to "Indicator" dashboard.
1. Click on "share" icon in the top right corner.
1. Select export and save file to [dashboard.json](./dashboard.json)

## Update 

Run "udpate grafana dashboard" CI job.
