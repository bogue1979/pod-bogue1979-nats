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
