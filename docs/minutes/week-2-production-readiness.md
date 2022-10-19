# 05/10 : Second production-readiness team meeting


## Participants : 

Production-readiness team (Amine Benaziz, Albert Troussard), Noémien Kocher, Pierluca Borsò-Tan

## Questions pour presentation :

est ce qu’on va tous a tour de role expliquer l’implementation ?

on a un tps pour faire pres

ex on a repere x issues et corrige, des aspects de code coverage, des metrics, 

## Questions :

how often are created blocks: 

max 10 sec, si flood de transactions elles sont pas ttes inclues dans le bloc et les autres noeuds soupconnent le leader de misbehavior cad refuse some transactions.  ds ce cas le leader chnage. Dans DELA, le changement de leader se fait dans l'ordre des noeuds et pas sous forme d'election.

transactions sont prises ds l’ordre

it is proof of authority system cad transaction passe si 2/3 +1 blocks d’accord (mais du coup problemes si le nombre de nodes est trop petit)

On peut avoir autant de transactions qu’on veut, chaque block en traite autant qu’il peut et c'est la pool qui gere l’ordre des transactions.

La pool est partagee et donc publique a tous les nodes.


## Outils

### -*Sonar* : static analysis tool

si on va ds issues il explique pq x est une issue

code smells : un truc qui pue dans le code peut etre des conventions non respectees, des bouts de code mal ecrit ...etc
 
sonar → measures → project overview -> aspect general

activity et voir evolutions des lignes dupliquees …etc

### - *Prometheus* :

prometheus : permet de capturer des metrics qd le syst fonctionne, permet de recolter des donnees.

on peut utiliser prometheus pour par exp avoir des histogramme des requests ..etc

exp :

promelectionstatus dans mod.go

et l’appeller dans evoting.go


## a faire dans les prochaines semaines : 

-Preparer presentation

-code coverage

-reduire complexite cognitive de certaines fonctions

-augmenter et ameliorer commentaires et documentation

-trouver des petites taches mineures, corriger des petits bugs, essayer d’ecrire des integration test pour sse remettre a coder

-faire des pr






