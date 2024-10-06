import React, { useState } from "react";

interface LogFormProps {}

const LogForm: React.FC<LogFormProps> = () => {
  const [address, setAddress] = useState<string>("");
  const [port, setPort] = useState<string>("");
  const [protocol, setProtocol] = useState<string>("tcp");
  const [messages, setMessages] = useState<string[]>([""]);
  const [response, setResponse] = useState<{ text: string; type: string } | null>(null);
  
  const [isAdvancedOptionsOpen, setIsAdvancedOptionsOpen] = useState<boolean>(false);
  const [facility, setFacility] = useState<number>(1);
  const [severity, setSeverity] = useState<number>(5);
  const [interval, setInterval] = useState<number>(5);

  const isValidAddress = (address: string) => {
    const ipPattern = /^(localhost|127\.0\.0\.1|\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b)$/;
    return ipPattern.test(address);
  };

  const isValidPort = (port: string) => {
    const portNum = Number(port);
    return portNum >= 1 && portNum <= 65535;
  };

  const handleAddMessage = () => setMessages((prev) => [...prev, ""]);

  const handleMessageChange = (index: number, value: string) => {
    setMessages((prev) => {
      const newMessages = [...prev];
      newMessages[index] = value;
      return newMessages;
    });
  };

  const handleRemoveMessage = (index: number) => {
    setMessages((prev) => prev.filter((_, i) => i !== index));
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const logConfig = { address, port, protocol, messages, facility, severity, interval };

    if (!isValidAddress(address)) return setResponse({ text: "Dirección IP no es válida.", type: "error" });
    if (!isValidPort(port)) return setResponse({ text: "Puerto debe estar entre 1 y 65535.", type: "error" });

    try {
      const res = await fetch("http://localhost:8080/sendlog", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(logConfig),
      });

      if (!res.ok) throw new Error("Error en la respuesta del servidor");

      const data = await res.json();
      setResponse({ text: data.status || "Log enviado con éxito.", type: "info" });
    } catch (error) {
      console.error("Error sending log:", error);
      setResponse({ text: "Error enviando log.", type: "error" });
    }
  };

  const getResponseClass = (type: string) => {
    const classes: { [key: string]: string } = {
      info: "bg-blue-100 text-blue-800 border-blue-300",
      error: "bg-red-100 text-red-800 border-red-300",
    };
    return classes[type] || "";
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-100">
      <div className="max-w-lg w-full p-6 bg-white shadow-md rounded-lg">
        {response && (
          <p className={`mb-4 text-center p-2 rounded border ${getResponseClass(response.type)}`}>
            {response.text}
          </p>
        )}
        <h2 className="text-2xl font-bold text-center mb-4">Enviar Logs</h2>
        <form onSubmit={handleSubmit}>
          <InputField label="Dirección IP" value={address} onChange={setAddress} type="text" required />
          <InputField label="Puerto" value={port} onChange={setPort} type="number" required />
          <SelectField label="Protocolo" value={protocol} onChange={setProtocol} options={["tcp", "udp"]} />

          <div className="mb-4">
            <button type="button" onClick={() => setIsAdvancedOptionsOpen(!isAdvancedOptionsOpen)} className="w-full bg-gray-300 text-gray-800 p-2 rounded flex items-center justify-between hover:bg-gray-400 transition">
              <span>{isAdvancedOptionsOpen ? "Ocultar Opciones Avanzadas" : "Mostrar Opciones Avanzadas"}</span>
            </button>
            {isAdvancedOptionsOpen && (
              <AdvancedOptions
                facility={facility}
                setFacility={setFacility}
                severity={severity}
                setSeverity={setSeverity}
                interval={interval}
                setInterval={setInterval}
              />
            )}
          </div>

          <MessagesSection messages={messages} onChange={handleMessageChange} onAdd={handleAddMessage} onRemove={handleRemoveMessage} />

          <button type="submit" className="w-full bg-green-500 text-white p-2 rounded hover:bg-green-600 transition">Enviar Logs</button>
        </form>
      </div>
    </div>
  );
};

const InputField: React.FC<{ label: string; value: string | number; onChange: (value: string) => void; type: string; required?: boolean }> = ({ label, value, onChange, type, required }) => (
  <div className="mb-4">
    <label className="block mb-1 text-gray-700">{label}</label>
    <input
      type={type}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="w-full border border-gray-300 p-2 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
      required={required}
    />
  </div>
);

const SelectField: React.FC<{ label: string; value: string; onChange: (value: string) => void; options: string[] }> = ({ label, value, onChange, options }) => (
  <div className="mb-4">
    <label className="block mb-1 text-gray-700">{label}</label>
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="w-full border border-gray-300 p-2 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
    >
      {options.map((option) => (
        <option key={option} value={option}>{option.toUpperCase()}</option>
      ))}
    </select>
  </div>
);

const AdvancedOptions: React.FC<{ facility: number; setFacility: (value: number) => void; severity: number; setSeverity: (value: number) => void; interval: number; setInterval: (value: number) => void; }> = ({ facility, setFacility, severity, setSeverity, interval, setInterval }) => (
  <div className="mt-4 p-4 border border-gray-300 rounded bg-gray-50">
    <InputField label="Facility" value={facility} onChange={(value) => setFacility(Number(value))} type="number" />
    <InputField label="Severity" value={severity} onChange={(value) => setSeverity(Number(value))} type="number" />
    <InputField label="Interval (segundos)" value={interval} onChange={(value) => setInterval(Number(value))} type="number" />
  </div>
);

const MessagesSection: React.FC<{ messages: string[]; onChange: (index: number, value: string) => void; onAdd: () => void; onRemove: (index: number) => void; }> = ({ messages, onChange, onAdd, onRemove }) => (
  <div className="mb-4">
    <label className="block mb-1 text-gray-700">Mensajes</label>
    {messages.map((message, index) => (
      <div key={index} className="mb-2 flex items-center">
        <input
          type="text"
          value={message}
          onChange={(e) => onChange(index, e.target.value)}
          className="flex-grow border border-gray-300 p-2 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
          required
        />
        <button type="button" onClick={onAdd} className="ml-2 bg-blue-500 text-white p-2 rounded hover:bg-blue-600 transition">+</button>
        {messages.length > 1 && (
          <button type="button" onClick={() => onRemove(index)} className="ml-2 bg-red-500 text-white p-2 rounded hover:bg-red-600 transition">-</button>
        )}
      </div>
    ))}
  </div>
);

export default LogForm;
