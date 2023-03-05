import sqlite3 from 'better-sqlite3';
const db = new sqlite3('/Users/theodub/Desktop/Forum/usersForum.db');


// LIKES 
export function likes(postId) {
    let buttonLike = document.querySelector(".like");
    let spanLike = document.querySelector(".like-count");
    buttonLike.addEventListener("click", () => {
        // Enregistrement du like dans la base de données
        const insert = db.prepare('INSERT INTO likes (post_id) VALUES (?)');
        insert.run(postId);

        // Mise à jour du nombre de likes dans le document HTML
        const count = db.prepare('SELECT COUNT(*) AS count FROM likes WHERE post_id = ?').get(postId);
        spanLike.innerHTML = count.count;
    });
}







// DISLIKES
      // function dislike() {
      //   let buttonLike = document.querySelector(".dislike")
      //   let spanLike = document.querySelector(".dislike-count")

      //   buttonLike.addEventListener("button", () => {
      //     likes
      //   })
      //   spanLike.addEventListener("span", () => {
      //     likes
      //   })
      //   let likeButton = document.querySelector('.dislike');
      //   let likeCount = document.querySelector('.dislike-count');
      //   let like = 0;
      //   likeButton.addEventListener("click", () => {
      //     like++;
      //     likeCount.innerHTML = like;

      //     console.log(like)
      //   });
      // }
      // dislike()