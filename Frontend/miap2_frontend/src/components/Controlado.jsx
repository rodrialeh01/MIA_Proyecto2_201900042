import { useState } from "react";

const Controlado = () => {

    const [todo, setTodo] = useState({
        title: "Todo #01",
        description: "Descripcion #01",
        state: "pendiente"
    });

    const handleSubmit = (e) => {
        e.preventDefault();
    }
    
    const handleChange = e => {
        console.log(e.target.value);
        console.log(e.target.name);
        setTodo({
            ...todo,
            [e.target.name]: e.target.value
        })
    }

    return (
        <form onSubmit={handleSubmit} >
            <input 
                type="text" 
                placeholder="Ingrese Todo" 
                className="form-control mb-2"
                name="title"
                value={todo.title}
                onChange={handleChange}
            />
            <textarea 
                className="form-control mb-2" 
                placeholder="Ingrese Descripcion"
                name="description"
                value={todo.description}
                onChange={handleChange}
            />
            <select className="form-select mb-2" name="state" value={todo.state} onChange={handleChange}>
                <option value="pendiente">Pendiente</option>
                <option value="completado">Completado</option>
            </select>
            <button type="submit" className="btn btn-primary">
                Procesar
            </button>
        </form>
    );
}

export default Controlado;