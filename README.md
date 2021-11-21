# cale

`cale` is a command-line utility for summarizing Calendly availability.

# getting started

1. `go install github.com/btc/cale`
1. cale hits the Calendly API and requires authentication. Set the value of the `CALENDLY_TOKEN` environment variable to a [Calendly personal access token](https://calendly.com/integrations/api_webhooks). Optionally, you may write the environment variable to a dotenv file in the home directory at `~/.env`.

# usage

Here, I summarize availability from a Calendly event named `60m`:
```
λ. cale 60m

Sun 21 Nov    from 12:00 PM to 3:00 PM
Mon 22 Nov    02:30 PM
Mon 22 Nov    from 5:30 PM to 7:30 PM
```
---

Let's see weekdays only:
```
λ. cale 60m -w

Mon 22 Nov    02:30 PM
Mon 22 Nov    from 5:30 PM to 7:30 PM
```
---

Now, let's exclude slots that end after 6:30 PM. 

```
λ. cale 60m -w -e 6:30PM

Sun 21 Nov    from 12:00 PM to 3:00 PM
Mon 22 Nov    02:30 PM
Mon 22 Nov    05:30 PM
```
💡 Note that the output displays time ranges (i.e. `from 12:00 PM to 3:00 PM`) only when useful. Otherwise, availability is simply represented by start time (e.g. `5:30 PM`).

---

Let's see up to 21 days into the future:
```
λ. cale 60m -w -e 6:30PM -n 21

Mon 22 Nov    02:30 PM
Mon 22 Nov    05:30 PM
Wed 01 Dec    from 3:30 PM to 6:30 PM
Thu 02 Dec    from 2:00 PM to 6:30 PM
Fri 03 Dec    from 2:00 PM to 3:30 PM
Fri 03 Dec    from 5:00 PM to 6:30 PM
```
