# pod-bogue1979-nats

A [pod](https://github.com/babashka/babashka.pods) to interact with [NATS](https://docs.nats.io/) using [babashka](https://github.com/borkdude/babashka/).

## kv
```clojure
(require '[babashka.pods :as pods])
(pods/load-pod "pod-bogue1979-nats")
(require '[pod.bogue1979.nats :as nats])

(def opts
  {:host "localhost"
   :nkey (slurp (str (System/getenv "HOME") "/.config/nats/local.seed")),
   :msg "_"
   :bucket "first-bucket"})

(prn (nats/kvput (merge opts {:key "foo" :value "bar"})))
(prn (nats/kvget (merge opts {:key "foo"})))

(nats/kvwatchbucket (fn [msg] (prn msg)) opts)
(deref (promise))
```

## publish
```clojure
(require '[babashka.pods :as pods])
(pods/load-pod "pod-bogue1979-nats")
;(some? (find-ns 'pod.bogue1979.nats))
(require '[pod.bogue1979.nats :as nats])

(def opts
  {:host "localhost"
   :nkey (slurp (str (System/getenv "HOME") "/.config/nats/local.seed")),
   :msg "{\"foo\": \"baz\"}"
   :subject "metrics.test"})

(prn (nats/publish opts))
;(prn (nats/publish (merge opts {:subject "not.allowed"})))
```

## subscribe
```clojure
(require '[babashka.pods :as pods]
         '[cheshire.core :as json :refer [parse-string]])
(pods/load-pod "pod-bogue1979-nats")
(require '[pod.bogue1979.nats :as nats])

(def opts
  {:host "localhost",
   :nkey (slurp (str (System/getenv "HOME") "/.config/nats/local.seed")),
   :subject "metrics.test"})

(defn prn-message
  [msg]
  (prn (try (parse-string (:data msg) true)
            (catch Exception _ {:error "Error parsing json", :message msg})
            (finally (:data msg)))))

(nats/subscribe (fn [x] (prn-message x)) opts)
(deref (promise))
```

## request
```bash
nats reply metrics.test "answer"
```

```clojure
(require '[babashka.pods :as pods])
(pods/load-pod "pod-bogue1979-nats")
(require '[pod.bogue1979.nats :as nats])

(def opts
  {:host "localhost"
   :nkey (slurp (str (System/getenv "HOME") "/.config/nats/local.seed")),
   :subject "metrics.test"
   :msg "request"
   :timeout_seconds 10})

(println (try
           (nats/request opts)
           (catch Exception e (str "Error: " (.getMessage e)))))
```
