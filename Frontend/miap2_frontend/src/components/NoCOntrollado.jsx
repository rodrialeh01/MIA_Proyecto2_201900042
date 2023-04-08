import { useRef } from "react";

const NoControlado = () => {

    const form = useRef(null)

    const handleSubmit = (e) => {
        e.preventDefault();

        const data = new FormData(form.current);

        const {title, description, state} = Object.fromEntries([...data.entries()]);
        console.log(title, description, state);
    }
     

    return (
        <form onSubmit={handleSubmit} ref={form}>
            <input 
                type="text" 
                placeholder="Ingrese Todo" 
                className="form-control mb-2"
                name="title"
            />
            <textarea 
                className="form-control mb-2" 
                placeholder="Ingrese Descripcion"
                name="description"
            />
            <select className="form-select mb-2" name="state">
                <option value="pendiente">Pendiente</option>
                <option value="completado">Completado</option>
            </select>
            <button type="submit" className="btn btn-primary">
                Procesar
            </button>
        </form>
    );
}

export default NoControlado;
