export const sendLog = async (logConfig: any) => {
    try {
      const res = await fetch("http://localhost:8080/sendlog", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(logConfig),
      });
  
      if (!res.ok) {
        const errorData = await res.json();
        throw new Error(errorData.message || "Error en la respuesta del servidor");
      }
  
      const data = await res.json();
      return data;
    } catch (error) {
      console.error("Error sending log:", error);
      const errorMessage = error instanceof Error ? error.message : "Error enviando log.";
      throw new Error(errorMessage);
    }
  };