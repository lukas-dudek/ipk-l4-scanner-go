# IPK Projekt 1 – OMEGA: Síťový L4 Scanner

**Kurz:** IPK 2025/2026
**Autor:** Lukáš Dudek (xdudekl00)

## Představení projektu

Vytvořil jsem síťový scanner pro čtvrtou vrstvu (L4). Program je napsaný v jazyce Go a dokáže skenovat TCP a UDP porty. Podporuje jak moderní IPv6, tak klasické IPv4. Výsledkem skenování je přehledný výpis stavu portů (otevřený, uzavřený nebo filtrovaný), přesně podle zadání.

## Funkce

Při implementaci jsem se snažil o to, aby byl scanner rychlý a spolehlivý.

1. **TCP Skenování**: Používám metodu TCP SYN scan (tzv. half-open). Funguje to tak, že se pošle SYN paket a čeká se na odpověď. Pokud přijde SYN-ACK, port je otevřený. Pokud RST, je zavřený. Když nepřijde nic, považuji ho za filtrovaný. Výhodou je, že se nedokončuje celý 3-way handshake, což je rychlejší.
2. **UDP Skenování**: U UDP je to trochu složitější, protože protokol je bezestavový. Posílám prázdný UDP datagram a vlastně čekám, jestli se mi nevrátí "chybová" hláška ICMP Destination Unreachable. Pokud se vrátí (Type 3 Code 3 pro IPv4 nebo Type 1 Code 4 pro IPv6), port je zavřený. V ostatních případech je otevřený (podle RFC 1122).
3. **Podpora IPv6**: Program plnohodnotně zvládá obě verze IP. Automaticky pozná, co skenuje, a podle toho sestaví správné hlavičky a kontrolní součty.

## Návrh

- **Raw sockety místo pcap zápisu**: Rozhodl jsem se použít raw sockety (`net.ListenPacket`). Je to jednodušší než ručně skládat celé Ethernet rámce a řešit ARP tabulky. Tímhle způsobem nechám operační systém, aby se postaral o linkovou vrstvu, což funguje spolehlivě. Tento způsob fungoval při vývoji na macOS.
- **Odchytávání odpovědí přes pcap**: Odpovědi chytám pomocí knihovny `pcap`. Nastavil jsem BPF filtry, tak aby bral jen to, co se týká potřebných portů.
- **Sdílení socketů**: Abych neplýtval prostředky systému, neotevírám pro každé vlákno nový socket. Mám čtyři centrální (v4/v6 pro TCP a UDP), které si všechny skenovací gorutiny sdílejí.
- **Rychlost vs. stabilita**: Skenování běží paralelně (gorutiny). Abych ale neodrovnal síťovku nebo abych neztrácel pakety, omezil jsem počet najednou běžících skenů na 50 a přidal jsem drobné zpoždění (`20ms`) mezi starty. Tohle může působit zvláštně, ale poradil mi to kolega co byl s projektem napřed.

## Spuštění

Program potřebuje práva rootu (`sudo`), protože raw sockety a pcap prostě z uživatelského režimu nefungují. Běží to v prostředí `nix`.

1. **Vstoupit do nix shellu**:

   ```bash
   nix develop "git+https://git.fit.vutbr.cz/NESFIT/dev-envs.git#go"
   ```
2. **Přeložit projekt**:

   ```bash
   make build
   ```
3. **Spuštění**:

   ```bash
   sudo ./ipk-L4-scan -i <INTERFACE> [-u PORTS] [-t PORTS] HOST [-w TIMEOUT]
   ```

### Pár příkladů:

- `sudo ./ipk-L4-scan -i eth0 -t 22,80-100 localhost` (Sken vybraných TCP portů)
- `sudo ./ipk-L4-scan -i eth0 -u 53 8.8.8.8` (Sken UDP portu 53)
- `./ipk-L4-scan -i` (Stačí pro výpis dostupných rozhraní)

## Známá omezení

O žádných omezeních nevím. Všechno, co bylo v zadání, by mělo fungovat.

## Testování

**Obsah testů byl vygenerován pomocí Gemini Pro 3.1 a průběh implementace byl konzultován s Gemini Pro 3.1.**

Snažil jsem se, aby testy nebyly jen pro formu, ale aby něco ověřovaly. Celkem jich tam mám 50.

### Automatizované testy (`make test`)

Testuji hlavně logiku, co se děje dřív, než se vůbec dotkneme sítě:

- **Argumenty**: Jestli to dobře parsuje rozsahy typu `1-1000`, jestli to hlídá nesmysly jako port `65536` nebo jestli to zvládne IPv6 adresy v různých formátech.
- **DNS**: Jestli se dobře překládají domény i adresy.
- **Pakety**: Jestli skládám správně TCP a UDP hlavičky (hlavně ty kontrolní součty pro v4 a v6).

Spustíte je: `make test`.

### Ruční testování (Integration)

Ověřoval jsem si to i v praxi na localhostu:

- Pustil jsem si dva testovací servery (přes Python na portech 8080 a 8081).
- Zkusil jsem skenovat `127.0.0.1` i `::1`.
- Všechny stavy (`open` u serverů, `closed` u prázdných portů, `filtered` při simulaci timeoutu) se vypisovaly přesně tak, jak mají.

## Použité zdroje

- RFC 793 a RFC 1122
- Dokumentace gopacket
- Dokumentace pflag
- Google Gemini Pro 3.1
# ipk-l4-scanner-go
