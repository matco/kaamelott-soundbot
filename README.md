# Kaamelott soundbot
This is a Slack command for the [Kaamelott soundboard](https://kaamelott-soundboard.2ec0b4.fr).

## Available commands
* search (or s)
* play (or p)
* random (or r)
* help (or h)

## Examples
Search for a sound:
```
/kaamelott search sais
```

You will get a list of sounds that match your search:
```
4: Ah bah ça après, moi, pour le détail, je sais pas
236: Mais si on fait marrer tout le monde avec nos chenilles à la purée de fraise, nos couilles d'oursin aux amandes, j'sais pas quelles autres saloperies là.
263: J'vous disais que j'étais victime des colifichés et qu'il faudrait qu'on commence à me considérer en tant que tel.
281: Bah je sais pas ? Me lâcher la grappe par exemple ?
```

Add a link to the sound in the current channel:
```
/kaamelott play 4
```

The following text will be posted:
```
https://kaamelott-soundboard.2ec0b4.fr/#son/apres_pour_le_detail_je_sais_pas
```

Add a link to a random sound in the current channel:
```
/kaamelott random
```

## Setup on your Slack
Go to your Slack configuration page ```https://yourdomain.slack.com/apps/manage/custom-integrations```. Choose ```Slack commands```, then ```Add Configuration```.
In the configuration page:
* For ```Command```, choose the text you want to type use to invoke the command (/kaamelott seems to be a good match).
* For ```URL```, enter the URL of your own instance or my instance ```https://kaamelott-soundbot.projects.matco.name```.
* For ```Method``` choose "POST".
* We don't care about the token.
* For the rest, it's up to you, but I suggest ```search | s <query>, play | p <id>, random | r``` for the usage hint.

## Install your own instance on Google App Engine
Checkout project from Github:
```
git clone git@github.com:matco/kaamelott-soundbot.git
```

Deploy to Google App Engine:
```
cd kaamelott-soundbot
gcloud app deploy app.yaml
```

You're done!

### Test the command
If you would like to test your instance, here are some sample requests that will be sent by Slack:
```
curl -d "text=search%20sais" -H "Content-Type: application/x-www-form-urlencoded" -X POST http://your.instance
curl -d "text=play%204" -H "Content-Type: application/x-www-form-urlencoded" -X POST http://your.instance
```
