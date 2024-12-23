(require '[babashka.pods :as pods]
         '[cheshire.core :as json :refer [parse-string]])
(pods/load-pod "pod-bogue1979-nats")
(require '[pod.bogue1979.nats :as nats])

(def opts
  {:host "localhost",
   :nkey (slurp "/home/danny/.config/nats/local.seed"),
   :subject "metrics.test"})

(defn prn-message
  [msg]
  (prn (try (parse-string (:data msg) true)
            (catch Exception _ {:error "Error parsing json", :message msg})
            (finally (:data msg)))))

(nats/subscribe (fn [x] (prn-message x)) opts)
(deref (promise))
