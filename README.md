API available at https://power.ffail.win/?zone=NO2&date=2021-06-17

Domains:
- NO1: 10YNO-1--------2
- NO2: 10YNO-2--------T
- NO3: 10YNO-3--------J
- NO4: 10YNO-4--------9
- NO5: 10Y1001A1001A48H

XML version from Entsoe:
```bash
YEAR=2025 MONTH=01 DAY=22 ; curl "https://web-api.tp.entsoe.eu/api?documentType=A44&in_Domain=10YNO-2--------T&out_Domain=10YNO-2--------T&periodStart=${YEAR}${MONTH}$((DAY-1))2300&periodEnd=${YEAR}${MONTH}${DAY}2300&securityToken=$(op item get entsoe.eu --fields 'Web Api Security Token')" > ~/tmp/60m.xml
```
