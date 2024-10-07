import React, { useReducer } from "react";
import { FaCog, FaExclamationCircle } from "react-icons/fa";
import { toast } from "react-toastify";
import InputField from "./InputField";
import SelectField from "./SelectField";
import MessagesSection from "./MessagesSection";
import AdvancedOptionsModal from "./AdvancedOptionsModal";
import { sendLog } from "../utils/sendLog";

interface LogFormProps {}

const initialState = {
  address: "",
  port: "",
  protocol: "tcp",
  messages: [""],
  facility: 1,
  severity: 5,
  interval: 5,
  isAdvancedOptionsOpen: false,
  errors: {
    address: "",
    port: "",
    messages: [""],
  },
};

type State = typeof initialState;

type Action =
  | { type: "SET_FIELD"; field: string; value: any }
  | { type: "ADD_MESSAGE" }
  | { type: "REMOVE_MESSAGE"; index: number }
  | { type: "SET_ERROR"; field: string; value: string }
  | { type: "SET_MESSAGE_ERROR"; index: number; value: string }
  | { type: "TOGGLE_ADVANCED_OPTIONS" };

const reducer = (state: State, action: Action): State => {
  switch (action.type) {
    case "SET_FIELD":
      return { ...state, [action.field]: action.value };
    case "ADD_MESSAGE":
      return { ...state, messages: [...state.messages, ""], errors: { ...state.errors, messages: [...state.errors.messages, ""] } };
    case "REMOVE_MESSAGE":
      return {
        ...state,
        messages: state.messages.filter((_, i) => i !== action.index),
        errors: { ...state.errors, messages: state.errors.messages.filter((_, i) => i !== action.index) },
      };
    case "SET_ERROR":
      return { ...state, errors: { ...state.errors, [action.field]: action.value } };
    case "SET_MESSAGE_ERROR":
      return {
        ...state,
        errors: {
          ...state.errors,
          messages: state.errors.messages.map((error, i) => (i === action.index ? action.value : error)),
        },
      };
    case "TOGGLE_ADVANCED_OPTIONS":
      return { ...state, isAdvancedOptionsOpen: !state.isAdvancedOptionsOpen };
    default:
      return state;
  }
};

const LogForm: React.FC<LogFormProps> = () => {
  const [state, dispatch] = useReducer(reducer, initialState);

  const isValidAddress = (address: string) => {
    const ipPattern = /^(localhost|127\.0\.0\.1|\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b)$/;
    return ipPattern.test(address);
  };

  const isValidPort = (port: string) => {
    const portNum = Number(port);
    return portNum >= 1 && portNum <= 65535;
  };

  const handleAddMessage = () => dispatch({ type: "ADD_MESSAGE" });

  const handleMessageChange = (index: number, value: string) => {
    dispatch({ type: "SET_FIELD", field: "messages", value: state.messages.map((msg, i) => (i === index ? value : msg)) });
    if (value.trim() === "") {
      dispatch({ type: "SET_MESSAGE_ERROR", index, value: "El mensaje no puede estar vacío." });
    } else {
      dispatch({ type: "SET_MESSAGE_ERROR", index, value: "" });
    }
  };

  const handleRemoveMessage = (index: number) => dispatch({ type: "REMOVE_MESSAGE", index });

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const { address, port, protocol, messages, facility, severity, interval } = state;
    const logConfig = { address, port, protocol, messages, facility, severity, interval };

    if (!isValidAddress(address)) {
      dispatch({ type: "SET_ERROR", field: "address", value: "Dirección IP no es válida." });
      return;
    }
    if (!isValidPort(port)) {
      dispatch({ type: "SET_ERROR", field: "port", value: "Puerto debe estar entre 1 y 65535." });
      return;
    }
    if (messages.some(message => message.trim() === "")) {
      toast.error(<><FaExclamationCircle /> Todos los mensajes deben contener texto.</>, { autoClose: 5000 });
      return;
    }

    try {
      const data = await sendLog(logConfig);
      toast.info(<> {data.status || "Log enviado con éxito."}</>, { autoClose: 5000 });
    } catch (error) {
      toast.error(<> {(error as Error).message}</>, { autoClose: 5000 });
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-100">
      <div className="max-w-lg w-full p-6 bg-white shadow-md rounded-lg relative">
        <h2 className="text-2xl font-bold text-center mb-4">Enviar Logs</h2>
        <form onSubmit={handleSubmit}>
          <InputField
            label="Dirección IP"
            value={state.address}
            onChange={(value) => {
              dispatch({ type: "SET_FIELD", field: "address", value });
              if (!isValidAddress(value)) {
                dispatch({ type: "SET_ERROR", field: "address", value: "Dirección IP no es válida." });
              } else {
                dispatch({ type: "SET_ERROR", field: "address", value: "" });
              }
            }}
            type="text"
            required
            error={state.errors.address}
          />
          <InputField
            label="Puerto"
            value={state.port}
            onChange={(value) => {
              dispatch({ type: "SET_FIELD", field: "port", value });
              if (!isValidPort(value)) {
                dispatch({ type: "SET_ERROR", field: "port", value: "Puerto debe estar entre 1 y 65535." });
              } else {
                dispatch({ type: "SET_ERROR", field: "port", value: "" });
              }
            }}
            type="number"
            required
            error={state.errors.port}
          />
          <SelectField label="Protocolo" value={state.protocol} onChange={(value) => dispatch({ type: "SET_FIELD", field: "protocol", value })} options={["tcp", "udp"]} />

          <MessagesSection messages={state.messages} onChange={handleMessageChange} onAdd={handleAddMessage} onRemove={handleRemoveMessage} errors={state.errors.messages} />

          <button type="submit" className="w-full bg-green-500 text-white p-2 rounded hover:bg-green-600 transition">Enviar Logs</button>
        </form>

        <button
          type="button"
          onClick={() => dispatch({ type: "TOGGLE_ADVANCED_OPTIONS" })}
          className="absolute top-4 right-4 text-gray-600 hover:text-gray-800 transition"
          aria-label="Opciones avanzadas"
        >
          <FaCog size={24} />
        </button>

        {state.isAdvancedOptionsOpen && (
          <AdvancedOptionsModal
            facility={state.facility}
            setFacility={(value) => dispatch({ type: "SET_FIELD", field: "facility", value })}
            severity={state.severity}
            setSeverity={(value) => dispatch({ type: "SET_FIELD", field: "severity", value })}
            interval={state.interval}
            setInterval={(value) => dispatch({ type: "SET_FIELD", field: "interval", value })}
            onClose={() => dispatch({ type: "TOGGLE_ADVANCED_OPTIONS" })}
          />
        )}
      </div>
    </div>
  );
};

export default LogForm;