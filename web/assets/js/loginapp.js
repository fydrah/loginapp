function kubeconfigDownload(id) {
   var a = document.body.appendChild(
       document.createElement("a")
   );
   a.download = "kubeconfig.yaml";
   a.href = "data:text/plain;charset=utf-8," + encodeURIComponent(document.getElementById(id).textContent);
   a.click();
};
