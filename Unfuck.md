# Unfuck

> An English version is not yet available at the time. You may try to translate the following italian

**LEGGI ATTENTAMENTE**

Se sei arrivato qui stai attento, è difficile perdere il salvataggio ma ci puoi sempre riuscire se sei stupido abbastanza.

Ci sono tre opzioni: `Manual save`, `Reset`, `Remove lock`.

## Manual Save

**Cosa fa**: salva tutti i cambiamenti che hai fatto al mondo e li pubblica su git

**Quando va usato**: se il programma è crashato in precedenza, non è stato chiuso correttamente o in ogni modo la sessione precedente non è stata salvata. Potresti voler usare `Remove lock` dopo.

## Reset

**Cosa fa**: sovrascrive qualsiasi cambiamento tu abbia fatto nella tua versione locale e recupera i file più recenti da git. In questo modo dovresti recuperare l'ultima versione funzionante del server

**Quando va usato**: se avete modificato per sbaglio dei file nella cartella del server e non vi interessa salvarli o avete fatto casino in qualche modo

## Remove lock

**Cosa fa**: rimuove il file che tiene traccia di chi ha aperto il server, effettivamente stai dicendo che sei _sicuro_ che nessuno ha il server aperto

**Quando va usato**: dopo il `Manual save` può darsi che il file di lock sia rimasto
