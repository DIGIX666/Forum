$(document).ready(function() {
    $('#myForm').submit(function(event) {
      // Empêche l'envoi du formulaire par défaut
      event.preventDefault();
      
      // Récupère les données du formulaire
      var formData = $(this).serialize();
      
      // Envoie des données avec une requête AJAX
      $.ajax({
        type: 'POST',
        url: '/', // endpoint pour traiter les données
        data: formData,
        success: function(response) {
          // Traitement de la réponse du serveur
          console.log(response);
        },
        error: function(xhr, status, error) {
          // Traitement de l'erreur
          console.log(error);
        }
      });
    });
  });
  



// import sqlite3 from '/home/student07/.cache/typescript/4.9/node_modules/@types/better-sqlite3'
// const db = new sqlite3('/Users/theodub/Desktop/Forum/usersForum.db')


// // LIKES 
// function likes(postId) {
//     let buttonLike = document.querySelector(".like")
//     let spanLike = document.querySelector(".like-count")
//     buttonLike.addEventListener("click", () => {
//         // Enregistrement du like dans la base de données
//         const insert = db.prepare('INSERT INTO likes (post_id) VALUES (?)')
//         insert.run(postId)

//         // Mise à jour du nombre de likes dans le document HTML
//         const count = db.prepare('SELECT COUNT(*) AS count FROM likes WHERE post_id = ?').get(postId)
//         spanLike.innerHTML = count.count;
//     });
// }

// let postId = document.querySelector(".postid");
// likes(postId);
