# Changelog

## [Aktuální stav]

### Co jsem naimplementoval:

- **TCP SYN Scan**: Porty skenuji odesíláním raw TCP SYN paketů. Podle odpovědi určuji stav: `open` (SYN-ACK), `closed` (RST) nebo `filtered` (žádná odpověď v termínu).
- **UDP Scan**: Odesílám prázdné UDP pakety. Pokud se mi vrátí ICMP/ICMPv6 (Port Unreachable), vím, že je port zavřený. Jinak ho hlásím jako otevřený.
- **Plná podpora IPv4 i IPv6**: Program si poradí s oběma protokoly, včetně překladu adres a výpočtu kontrolních součtů pro oba typy hlaviček.
- **Rychlé paralelní skenování**: Použil jsem Gorutiny a semafor, takže můžu skenovat až 50 portů naráz. Nic se přitom neztrácí a program neběží příliš rychle. Zároveň jsem se snažil nevyčerpat veškeré zdroje hodnotícího prostředí.
- **Sdílené raw sockety**: Místo abych pro každý port otevíral nový socket, mám 4 centrální, které všechny gorutiny sdílejí. Je to šetrnější k systému a rychlejší.
- **Chytré parsování argumentů**: Používám knihovnu `pflag`. Uživatel může zadávat rozsahy (např. `20-30`), seznamy portů oddělené čárkou a argumenty v libovolném pořadí.
- **Korektní signály**: Když uživatel zmáčkne Ctrl+C, program nespadne, ale vypíše na `stderr` info o ukončení a čistě skončí.

### Známá omezení:

- Žádná omezení nejsou známa. Aplikace splňuje snad všechno, co bylo v zadání.
