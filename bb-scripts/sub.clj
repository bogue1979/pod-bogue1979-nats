(ns sub
  (:require [clojure.core.async :as a]
            [babashka.pods :as pods]
            [cheshire.core :as json :refer [parse-string]]
            [pub :as pub]))

(pods/load-pod "pod-bogue1979-nats")
(require '[pod.bogue1979.nats :as nats])

(def opts
  {:host "localhost",
   :nkey "SUABUVFNUZOLGESMUFWGGA766NRZUDCJFDM2JIQSYNVP5AIPWTC4GBMAZY",
   :subject "metrics.qa-server"})

(defn publish [h n s m] (pub/pub h n s m))

(defn prn-message
  [msg]
  (prn (try (parse-string (get-in msg [:nats_message :data]) true)
            (catch Exception _ {:error "Error parsing json", :msg msg})
            (finally (get-in msg [:nats_message :data])))))

(defn response-to
  [msg subj]
  (println (str "Request Content: " (get-in msg [:nats_message :data])))
  (publish (:host opts) (:nkey opts) subj "Antwort"))

(defn runsubscriber
  [ch]
  (.addShutdownHook (Runtime/getRuntime)
                    (Thread. #(do (println "SHUTDOWN") (a/close! ch))))
  (println "Start subscriber")
  (a/go (nats/subscribe (fn [msg] (a/>! ch msg)) opts)))

(defn handle-messages
  [ch]
  (while true
    (let [msg (a/<!! ch)
          rply-to (get-in msg [:nats_message :reply])]
      (if (not (empty rply-to)) (response-to msg rply-to) (prn-message msg)))))

(defn -main
  []
  (let [ch (a/chan)]
    (runsubscriber ch)
    (handle-messages ch)))

(when (= *file* (System/getProperty "babashka.file"))
  (apply -main *command-line-args*))
