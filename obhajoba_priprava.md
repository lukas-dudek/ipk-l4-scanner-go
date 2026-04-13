# Příprava na obhajobu IPK Projektu 1 (L4 Scanner)

Tento dokument slouží jako tvoje příprava pro obhajobu podle zadání. Můžeš podle něj postupovat při prezentaci svého projektu.

## 1. Problém (Zadání)
Cílem projektu bylo vytvořit síťový scanner pro síťovou vrstvu L4 (TCP a UDP). Program měl být schopen přijmout argumenty příkazové řádky specifikující rozhraní, překládat doménová jména, umět pracovat jak s IPv4, tak s IPv6, a správně vyhodnotit stav skenovaných portů. Očekávanými stavy portů byly `open`, `closed` a případně `filtered` na základě odpovědí na vytvořené pakety. Muselo to fungovat svižně, zohledňovat timeouty a přesně odpovídat definici chování v RFC.

## 2. Naše řešení
Scanner jsem se rozhodl implementovat v jazyce **Go** pomocí knihovny **gopacket**.
- **Odesílání paketů:** Vyhnul jsem se složité konstrukci Ethernet rámců a MAC adres a odesílám pakety primárně pomocí Raw socketů pro L3 (na síťové vrstvě OS), které sestaví a pošlou IP hlavičku (pro správné checksumy v TCP/UDP jsem použil pseudo-hlavičky).
- **Zpracování odpovědí:** Odpovědi pak čtu čistě pomocí `libpcap` (gopacket/pcap) poslouchaného na zvoleném síťovém rozhraní s BPF (Berkeley Packet Filter), které je nastavené přesně na to, co potřebuji odchytit z cílové IP adresy a portů.
- **TCP (SYN scan / Half-open):** Posíláme pouze SYN paket. Jakmile zachytíme přes pcap `SYN-ACK`, tak cíl vyhodnotíme jako `open`. Když dostaneme `RST`, označíme ho jako `closed`. Pokud vyprší timeout, zaznamenáme `filtered`.
- **UDP scan:** UDP protokol je bezestavový, posílám prázdný UDP datagram a čekám na chybovou zprávu ICMP (ICMPv4 typ 3 kód 3 - Destination/Port Unreachable nebo ICMPv6 typ 1 kód 4). Pokud chybová zpráva přijde, vím, že port je `closed`. Pokud nevyjde ze serveru nic a vyprší timeout, uvažuji port jako `open` (ve shodě s RFC 1122), jelikož UDP nesignalizuje úspěch explicitně.
- **Konkurence a výkon:** Nasadil jsem Gorutiny. Otevřené listenery mám pouze 4 sdílené napříč pcap odchyty a do toho omezuji maximální počet běžících skenů na 50 s minimálním delayem (20ms), abych nazahltil síťový stack zařízení a neztrácel odpovědi.

## 3. Limitace
- **Práva uživatele:** Kvůli použití Raw socketů a `libpcap` je bezpodmínečně nutné scanner spouštět pod rootem (např. pomocí `sudo`). Z uživatelského režimu záchyt spojení nefunguje.
- **Nepřesnost UDP scanu:** Spoléháme se na ICMP odpovědi pro určení `closed` portu. Velkou limitací ale je, pokud je na cestě nastaven firewall pro „Dropping/Blackholing“ paketů (což je dnes běžné). Firewall nepodá žádnou ICMP hlášku, vyprší timeout a my port označíme jako `open`. Tohle je ale typická vlastnost protokolu UDP, se kterou scanner bez aplikačního kontextu moc nenadělá.
- **Vymezovací čas:** Pokud se skenuje hodně UDP portů a firewall pakety zahazuje, skenování potrvá opravdu dlouho, protože na každý neexistující UDP port logicky musí vypršet celý timeout.

## 4. Edge Cases (Okrajové případy)
- **Rozdíly IPv4 a IPv6:** Validace IP adresy nebo rozlišení, jestli má jít kontrolní součet s Ipv4 nebo Ipv6 pseudo-hlavičkou, a zároveň zjištění správného typu kódu pro ICMP Unreachable.
- **Doménová jména vracející víc IP adres:** Ve standardních TCP klientech se zkusí první a případně se volá další. Pro test vyřešeno, ale scanner má ujasněnou sémantiku.
- **Složité rozsahy portů:** Ošetřování nesmyslných hodnot (vysoká čísla nad 65535, zpětné pořadí atd.) řeším už během parsování a tyto skeny se ignorují nebo odhodí.
- **Konkurenční zachytávání v PCAP:** Sdílení file descriptorů. Sdílím sockety globálně, abych nenaboural limit maxima otevřených FD v operačním systému a nezahltil kernel.

## 5. Příprava opravy Makefile (Triviální úprava - PR)

**Příčina chyby na testech:**
Nástroj `make`, pokud nedostane specifický cíl jako doménu, spustí první cíl v `Makefile`. Z důvodu dřívějších úprav zůstal úplně jako první pravidlo `NixDevShellName` vypisující `@echo go`. Autotest proto zadal `make`, na stdout mu vyjelo jen `go` a nikdy nedošlo ke kompilaci – proto ten penalty koeficient za SingleHostNotCompiled.

**Přístup k opravě pro ukázku (Pull Request / Diff):**
- Přesunul jsem cíl `all: build` a `build:` na úplný začátek souboru, bezprostředně po direktivě `.PHONY`.
- Udělal jsem diff, na fakultní Giteu jej budeš demonstrovat jako Pull Request/Commit řešící ten chybějící defaultní stav (oprava na ~3 řádky, triviální změna pro chod celého projektu = 5 bodů penalizace, namísto 0 ze symetriky).

**Co tě čeká po obhajobě:**
Změnu zadokumentuješ do IS VUT novým zarchivováním souborů (s upraveným `Makefile`) jako novým submission podle instrukcí. Kód v gitu už je opraven. Do IS VUT se ale nesmí zapomenout udělat upload nového balíčku ZIP.
