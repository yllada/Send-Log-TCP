// src/components/LogForm.tsx
import React, { useState, useEffect } from "react";
import axios from "axios";

// Hook personalizado para manejar la configuración del log
const useLogConfig = () => {
  const [config, setConfig] = useState({
    address: "",
    protocol: "tcp",
    facility: 16,
    severity: 6,
    hostname: "",
    messages: "",
    interval: 5,
  });

  useEffect(() => {
    const savedConfig = localStorage.getItem("logConfig");
    if (savedConfig) {
      const parsedConfig = JSON.parse(savedConfig);
      setConfig((prevConfig) => ({
        ...prevConfig,
        ...parsedConfig,
      }));
    }
  }, []);

  const saveConfig = () => {
    localStorage.setItem("logConfig", JSON.stringify(config));
    alert("Configuración guardada exitosamente");
  };

  const clearConfig = () => {
    setConfig({
      address: "",
      protocol: "tcp",
      facility: 16,
      severity: 6,
      hostname: "",
      messages: "",
      interval: 5,
    });
    localStorage.removeItem("logConfig");
    alert("Campos limpiados y configuración borrada");
  };

  return { config, setConfig, saveConfig, clearConfig };
};

// Componente LogForm
const LogForm: React.FC = () => {
  const { config, setConfig, saveConfig, clearConfig } = useLogConfig();
  const [feedbackMessage, setFeedbackMessage] = useState("");
  const [isError, setIsError] = useState(false);
  const [loading, setLoading] = useState(false);
  const [fieldErrors, setFieldErrors] = useState<{ [key: string]: string }>({});

  // Validar los valores
  const isValid = (value: number, min: number, max: number) => {
    return value >= min && value <= max;
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setFieldErrors({}); // Reiniciar errores

    const { address, protocol, facility, severity, hostname, messages, interval } = config;

    // Validar campos antes de enviar
    const errors: { [key: string]: string } = {};
    if (!address) errors.address = "Address is required";
    if (!isValid(facility, 0, 23)) errors.facility = "Facility must be between 0 and 23";
    if (!isValid(severity, 0, 7)) errors.severity = "Severity must be between 0 and 7";
    if (!isValid(interval, 1, 60)) errors.interval = "Interval must be between 1 and 60 seconds";

    setFieldErrors(errors);
    if (Object.keys(errors).length > 0) return; // Detener si hay errores

    const logData = {
      address,
      protocol,
      facility,
      severity,
      hostname,
      messages: messages.split(","),
      interval,
    };

    setLoading(true); // Activar estado de carga
    try {
      const response = await axios.post(`${import.meta.env.VITE_API_URL}`, logData);
      setFeedbackMessage(`Log sent successfully! Server response: ${response.data.message || response.data}`);
      setIsError(false);
      resetForm();
    } catch (error) {
      const errorMessage =
        axios.isAxiosError(error) && error.response
          ? error.response.data.message || error.message
          : "Error sending log. Please try again.";
      setFeedbackMessage(errorMessage);
      setIsError(true);
    } finally {
      setLoading(false); // Desactivar estado de carga
    }
  };

  const resetForm = () => {
    setConfig({
      address: "",
      protocol: "tcp",
      facility: 16,
      severity: 6,
      hostname: "",
      messages: "",
      interval: 5,
    });
  };

  return (
    <div className="max-w-md mx-auto bg-white p-6 rounded-lg shadow-md">
      <h2 className="text-2xl font-bold mb-4">SendLog Syslog TCP/UDP</h2>

      {/* Mensaje de feedback */}
      {feedbackMessage && (
        <div className={`mb-4 p-2 rounded ${isError ? "bg-red-200 text-red-700" : "bg-green-200 text-green-700"}`}>
          {feedbackMessage}
        </div>
      )}

      <form onSubmit={handleSubmit} className="flex flex-col space-y-4">
        <div className="flex space-x-4">
          <div className="flex-1">
            <label className="block text-gray-700">Address</label>
            <input
              type="text"
              value={config.address}
              onChange={(e) => setConfig({ ...config, address: e.target.value })}
              required
              className={`mt-1 block w-full p-2 border border-gray-300 rounded ${fieldErrors.address ? "border-red-500" : ""}`}
            />
            {fieldErrors.address && <p className="text-red-500">{fieldErrors.address}</p>}
          </div>
          <div className="flex-1">
            <label className="block text-gray-700">Protocol</label>
            <select
              value={config.protocol}
              onChange={(e) => setConfig({ ...config, protocol: e.target.value })}
              className="mt-1 block w-full p-2 border border-gray-300 rounded"
            >
              <option value="tcp">TCP</option>
              <option value="udp">UDP</option>
            </select>
          </div>
        </div>

        <div className="flex space-x-4">
          <div className="flex-1">
            <label className="block text-gray-700">Facility</label>
            <input
              type="number"
              value={config.facility}
              onChange={(e) => setConfig({ ...config, facility: Number(e.target.value) })}
              required
              className={`mt-1 block w-full p-2 border border-gray-300 rounded ${fieldErrors.facility ? "border-red-500" : ""}`}
            />
            {fieldErrors.facility && <p className="text-red-500">{fieldErrors.facility}</p>}
          </div>
          <div className="flex-1">
            <label className="block text-gray-700">Severity</label>
            <input
              type="number"
              value={config.severity}
              onChange={(e) => setConfig({ ...config, severity: Number(e.target.value) })}
              required
              className={`mt-1 block w-full p-2 border border-gray-300 rounded ${fieldErrors.severity ? "border-red-500" : ""}`}
            />
            {fieldErrors.severity && <p className="text-red-500">{fieldErrors.severity}</p>}
          </div>
          <div className="flex-1">
            <label className="block text-gray-700">Hostname</label>
            <input
              type="text"
              value={config.hostname}
              onChange={(e) => setConfig({ ...config, hostname: e.target.value })}
              required
              className={`mt-1 block w-full p-2 border border-gray-300 rounded ${fieldErrors.hostname ? "border-red-500" : ""}`}
            />
            {fieldErrors.hostname && <p className="text-red-500">{fieldErrors.hostname}</p>}
          </div>
        </div>

        <div className="mb-4">
          <label className="block text-gray-700">Messages (comma separated)</label>
          <textarea
            value={config.messages}
            onChange={(e) => setConfig({ ...config, messages: e.target.value })}
            required
            className="mt-1 block w-full p-2 border border-gray-300 rounded"
          />
        </div>

        <div className="mb-4">
          <label className="block text-gray-700">Interval (in seconds)</label>
          <input
            type="number"
            value={config.interval}
            onChange={(e) => setConfig({ ...config, interval: Number(e.target.value) })}
            required
            className={`mt-1 block w-full p-2 border border-gray-300 rounded ${fieldErrors.interval ? "border-red-500" : ""}`}
          />
          {fieldErrors.interval && <p className="text-red-500">{fieldErrors.interval}</p>}
        </div>

        <div className="flex flex-wrap">
          <button
            type="button"
            onClick={saveConfig}
            className="bg-green-500 text-white p-2 m-2"
          >
            Save Config
          </button>
          <button
            type="button"
            onClick={clearConfig}
            className="bg-red-500 text-white p-2 m-2"
          >
            Clear Config
          </button>
          <button
            type="submit"
            disabled={loading} // Deshabilitar mientras se carga
            className={`bg-blue-500 text-white p-2 m-2 ${loading ? "opacity-50 cursor-not-allowed" : ""}`}
          >
            {loading ? "Sending..." : "Send Log"}
          </button>
        </div>
      </form>
    </div>
  );
};

export default LogForm;
