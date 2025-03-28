API available at https://power.ffail.win/?zone=NO2&date=2021-06-17

Domains:
- NO1: 10YNO-1--------2
- NO2: 10YNO-2--------T
- NO3: 10YNO-3--------J
- NO4: 10YNO-4--------9
- NO5: 10Y1001A1001A48H

XML version from Entsoe:
```bash
YEAR=2025 MONTH=1 DAY=8 ZONE=10YNO-2--------T d=0$DAY m=0$MONTH db=0$((DAY-1)); curl "https://web-api.tp.entsoe.eu/api?documentType=A44&in_Domain=${ZONE}&out_Domain=${ZONE}&periodStart=${YEAR}${m: -2}${db: -2}2300&periodEnd=${YEAR}${m: -2}${d: -2}2300&securityToken=$(op item get entsoe.eu --fields 'Web Api Security Token')"
```

Prod:
```bash
d=2025-03-08 z=NO2 ;curl "https://norway-power.ffail.win/?zone=${z}&date=${d}&key=$(op read op://Personal/power.ffail.win/api-key)" | jq
```
Staging:
```bash
d=2025-03-08 z=NO2 ;curl "https://latest---power-price-xvexnfx5sa-ew.a.run.app?zone=${z}&date=${d}&key=$(op read op://Personal/power.ffail.win/api-key)" | jq
```
