# Security

ClashBandwidthTest only communicates with the locally running Mihomo controller and the configured bandwidth test endpoint.

The application does not intentionally modify Windows System Proxy, TUN, DNS, WFP, routing tables, adapters, or registry proxy settings.

Do not expose Mihomo's external controller to untrusted networks. Prefer local IPC or a controller bound to `127.0.0.1`, and use a secret when TCP controller access is enabled.

Please avoid posting Clash/Mihomo controller secrets, subscription URLs, or provider credentials in public issues.
