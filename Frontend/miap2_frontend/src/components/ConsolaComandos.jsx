import React from 'react';
import Editor from '@monaco-editor/react';
import './ConsolaComandos.css';
import { useState, useRef } from 'react';
import Service from '../Services/Service';

const ConsolaComandos = () => {
    let archivo;

    const [comandos, setComandos] = useState('#Aqui puedes ingresar tus comandos');
    const [response, setResponse] = useState('DiskMIA - File Command System Console<br />Copyright (C) DiskMIA - File Command System Console MIA-P2. Created by Rodrigo Hernández 2023<br /><br />');
    const editorRef = useRef(null)


    const handleEditorDidMount = (editor, monaco) => {
        editorRef.current = editor;
    }
    
    const CargarArchivo = ( e )=> {
        archivo = e.target.files[0];
    }
    
    const handlerPostParse = () => {
        Service.parse(editorRef.current.getValue())
        .then(({respuesta})=>{
            console.log(respuesta)
            let res = 'DiskMIA - File Command System Console\nCopyright (C) DiskMIA - File Command System Console MIA-P2. Created by Rodrigo Hernández 2023\n\n'+respuesta
            setResponse(FontColorResponse(res))
        })
    }

    const ExaminarArchivo = () => {
        console.log(archivo)
        if(!archivo){
            alert('No se ha seleccionado ningun archivo');
            return;
        }
        const reader = new FileReader();
        reader.readAsText(archivo, 'UTF-8');
        reader.onload = () => {
            setComandos(reader.result);
            editorRef.current.setValue(reader.result);
        }
        reader.onerror = () => {
            alert('Error al leer el archivo');
        }
    }

    const FontColorResponse = (response) => {
        const regex_comentarios = /#.*$/gm;
        let text = response.replaceAll(regex_comentarios, '<span style="color: #06671B;" className="fuente1">$&</span>');
        const regex_error = /(\[-ERROR-\].*\n)/g;
        text = text.replaceAll(regex_error, '<span style="color: #ED0202;" className="fuente1">$&</span>');
        const regex_exito = /(\[\*SUCCESS\*].*\n)/g;
        text = text.replaceAll(regex_exito, '<span style="color: #33E100;" className="fuente1">$&</span>');
        const regex_warning = /(\[\/\\WARNING\/\\].*\n)/g;
        text = text.replaceAll(regex_warning, '<span style="color: #EDE600;" className="fuente1">$&</span>');
        text = text.replaceAll('!------------PARTICIONES MONTADAS------------!', '<span style="color: #007EE1;" className="fuente1">!------------PARTICIONES MONTADAS------------!</span>');
        text = text.replaceAll('!--------------------------------------------!', '<span style="color: #007EE1;" className="fuente1">!--------------------------------------------!</span>');
        const regex_id = /^(\t{2}ID: 42[0-9][A-Z])$/gm;
        text = text.replaceAll(regex_id, '<span style="color: #00E1D0;" className="fuente1">&emsp;&emsp;&emsp;$&</span>');
        const regex_ejecutando = /(\n\nEJECUTANDO: (mkdisk|fdisk|rmdisk|mount|mkfs|login|logout|mkgrp|rmgrp|mkusr|rmusr|mkfile|mkdir|rep).*\n)/g;
        text = text.replaceAll(regex_ejecutando, '<span style="color: #E18500;" className="fuente1">$&</span>');
        text = text.replaceAll('\n', '<br />');
        return text;
    }

    return (
    <>
    <h3 className='fuente1' style={{alignContent:'center', color:'white', marginLeft:'5%'}}>Cargar:</h3>
    <div style={{ display: 'flex', justifyContent: 'center'}}>
        <input class="form-control form-control-lg" id="formFileLg" style={{width:'90%'}}type="file" onChange={CargarArchivo} accept='.eea'/>
    </div>
    <br />
    <div style={{ display: 'flex', justifyContent: 'center' }}>
        <button style={{ marginLeft: 'auto', marginRight: 'auto', fontSize:'24px' }} className="btn colorbtn2" type="button" onClick={ExaminarArchivo}><svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-upload" viewBox="0 0 16 16">
  <path d="M.5 9.9a.5.5 0 0 1 .5.5v2.5a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1v-2.5a.5.5 0 0 1 1 0v2.5a2 2 0 0 1-2 2H2a2 2 0 0 1-2-2v-2.5a.5.5 0 0 1 .5-.5z"/>
  <path d="M7.646 1.146a.5.5 0 0 1 .708 0l3 3a.5.5 0 0 1-.708.708L8.5 2.707V11.5a.5.5 0 0 1-1 0V2.707L5.354 4.854a.5.5 0 1 1-.708-.708l3-3z"/>
</svg> Examinar</button>
    </div>
    <h3 className='fuente1' style={{justifyContent:'center', color:'white', marginLeft:'90px'}}>Comandos:</h3>
    <div style={{ display: 'flex', justifyContent: 'center'}}>
        <Editor
            size="50px"
            theme='vs-dark'
            fontSize='50px'
            language='python'
            options={{
                fontSize:16,
                lineNumbers:'off'
            }}
            width="90%"
            height="300px"
            value={comandos}
            onChange={(value) => setComandos(value)}
            onMount={handleEditorDidMount}
        />
    </div>
    <br />
    <div style={{ display: 'flex', justifyContent: 'right'}}>
        <button style={{ marginLeft: 'auto', marginRight: 'auto', fontSize:'24px'}} className="btn colorbtn1" type="button" onClick={handlerPostParse}><svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-send-check" viewBox="0 0 16 16">
        <path d="M15.964.686a.5.5 0 0 0-.65-.65L.767 5.855a.75.75 0 0 0-.124 1.329l4.995 3.178 1.531 2.406a.5.5 0 0 0 .844-.536L6.637 10.07l7.494-7.494-1.895 4.738a.5.5 0 1 0 .928.372l2.8-7Zm-2.54 1.183L5.93 9.363 1.591 6.602l11.833-4.733Z"/>
        <path d="M16 12.5a3.5 3.5 0 1 1-7 0 3.5 3.5 0 0 1 7 0Zm-1.993-1.679a.5.5 0 0 0-.686.172l-1.17 1.95-.547-.547a.5.5 0 0 0-.708.708l.774.773a.75.75 0 0 0 1.174-.144l1.335-2.226a.5.5 0 0 0-.172-.686Z"/>
        </svg> Ejecutar</button>
    </div>
    <h3 className='fuente1' style={{justifyContent:'center', color:'white', marginLeft:'5%'}}>Salida:</h3>
    <div style={{ display: 'flex', justifyContent: 'center'}}>
        <div class="form-floating" style={{ width:'90%', height:'500px'}}>
        <div readOnly className="form-control fuente2" dangerouslySetInnerHTML={{ __html: response }} style={{height:'500px', backgroundColor:'black', color:'white', overflowY: 'scroll'}}></div>
        </div>
    </div>
    </>
    ); 
}
export default ConsolaComandos;