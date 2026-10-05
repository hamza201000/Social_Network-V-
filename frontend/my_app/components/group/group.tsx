import { Group } from "@/types/groupData"
import { useEffect, useState } from "react"


export default function Groups({id}:{id:number}){
    const [group,setGroup]=useState<Group[]>([])
    const getGroup = async()=>{
        try{

            const response= await fetch(`http://localhost:8080/api/group/${id}`,{
                credentials:"include"
            }) 
            const data : Group[] = await response.json()
            setGroup(data)
            // if (!response.ok){
                
            // }
        }catch(err){

        }
    }
    useEffect(()=>{
        getGroup()
    },[id])
    return(
        <>
        
        </>
    )

}